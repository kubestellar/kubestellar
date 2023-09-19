#!/usr/bin/env python

# Copyright 2023 The KubeStellar Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from __future__ import annotations
import argparse
import git
import graphlib
import os
import stat
import time

class VertexData():
    commit: git.Commit
    
    def __init__(self, commit: git.Commit):
        self.commit = commit
    
    def __str__(self) -> str:
        return self.commit.hexsha
    
    def __lt__(self, other) -> bool:
        if self is other: return False
        if isinstance(other, VertexData): return self.commit.committed_date < other.commit.committed_date
        return NotImplemented
    
    pass

class Vertex():
    data: VertexData
    parents: set(Vertex)
    children: set(Vertex)
    ancestors: set(Vertex)
    
    def __init__(self, data: VertexData):
        self.data = data
        self.parents = set[Vertex]()
        self.children = set[Vertex]()
        self.ancestors = set[Vertex]()
    
    def add_parent(self, parent: Vertex) -> None:
        self.parents.add(parent)
        parent.children.add(self)
    
    def derive_ancestors(self) -> None:
        self.ancestors |= self.parents
        for parent in self.parents:
            self.ancestors |= parent.ancestors
        # print(f'Derived ancestors of {self.data}: count={len(self.ancestors)}')
        return
    
    def __lt__(self, other) -> bool:
        if self is other: return False
        if isinstance(other, Vertex): return self.data < other.data
        return NotImplemented
    
    def sorted_ancestors(self) -> list[Vertex]:
        ans = list(self.ancestors)
        ans.sort()
        return ans
    
    pass

VertexCache = dict[git.Commit, Vertex]

def get_vertex(commit: git.Commit, cache: VertexCache) -> Vertex:
    work = list[(git.Commit,Vertex)]()
    def inner_get(commit: git.Commit) -> Vertex:
        ans = cache.get(commit)
        if ans:
            return ans
        ans = Vertex(VertexData(commit))
        cache[commit] = ans
        for parent in commit.parents:
            work.append((parent, ans))
        # print(f'Queuing {ans.data}')
        return ans
    ans = inner_get(commit)
    while work:
        (parent_commit, child_vertex) = work[0]
        work = work[1:]
        parent_vertex = inner_get(parent_commit)
        child_vertex.add_parent(parent_vertex)
    return ans

def derive_ancestors(vc: VertexCache) -> None:
    sorter = graphlib.TopologicalSorter()
    for (commit, vertex) in vc.items():
        sorter.add(vertex, *vertex.parents)
    for vertex in sorter.static_order():
        vertex.derive_ancestors()
    return

def engraph_repo(repo: git.Repo) -> VertexCache:
    vc = VertexCache()
    for branch in repo.heads:
        get_vertex(branch.commit, vc)
    for tag in repo.tags:
        get_vertex(tag.commit, vc)
    derive_ancestors(vc)
    return vc

plainchars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+-_0123456789.'

def sanitize(instr: str) -> str:
    ans = ''
    for ch in instr:
        if ch in plainchars:
            ans = ans + ch
        else:
            n = ord(ch)
            ans = ans + '%{:04x}'.format(n)
    return ans

def write_ancestors(dir: str, vertex: Vertex) -> None:
    ancs = vertex.sorted_ancestors()
    filename = os.path.join(dir, str(vertex.data))
    with open(filename, 'w') as file:
        for anc in ancs:
            file.write(str(anc.data) + '\n')
    return

def ensure_symlink(src, dest) -> None:
    try:
        ls = os.lstat(dest)
    except:
        os.symlink(src, dest)
        return
    if stat.S_ISLNK(ls.st_mode):
        cur_src = os.readlink(dest)
        if cur_src == src:
            return
    os.remove(dest)
    os.symlink(src, dest)
    return

def write_indices(dir: str, repo: git.Repo, vc: VertexCache) -> None:
    commit_dir = os.path.join(dir, 'commits')
    tag_dir = os.path.join(dir, 'tags')
    branch_dir = os.path.join(dir, 'branches')
    remotebranch_dir = os.path.join(dir, 'remotebranches')
    rel_commit_dir = os.path.join('..', 'commits')
    os.makedirs(commit_dir, exist_ok=True)
    os.makedirs(tag_dir, exist_ok=True)
    os.makedirs(branch_dir, exist_ok=True)
    os.makedirs(remotebranch_dir, exist_ok=True)
    for ref in repo.references:
        if isinstance(ref, git.RemoteReference):
            parts = ref.name.split('/', 1)
            sym = os.path.join(remotebranch_dir, parts[0] + '-' + sanitize(parts[1]))
        elif isinstance(ref, git.Tag):
            sym = os.path.join(tag_dir, sanitize(ref.name))
        elif isinstance(ref, git.Head):
            sym = os.path.join(branch_dir, sanitize(ref.name))
        else:
            print('f{ref} has unexpected type {type(ref)}')
            continue
        vertex = get_vertex(ref.commit, vc)
        commit_filename = os.path.join(rel_commit_dir, str(vertex.data))
        ensure_symlink(commit_filename, sym)
    for (commit, vertex) in vc.items():
        write_ancestors(commit_dir, vertex)
    return

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Parse git metadata to files in index directories commits, tags, branches, remotebranches.')
    parser.add_argument('--repo', default='../..', help='directory holding .git, defaults to ../..')
    parser.add_argument('--output', default='..', help='directory that is parent of index directories, defaults to ..')
    args = parser.parse_args()
    repo = git.Repo(args.repo)
    vc = engraph_repo(repo)
    write_indices(args.output, repo, vc)
