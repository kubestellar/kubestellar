# Git metadata parser

This directory contains a git metadata parser, in
[scripts/index.py](scripts/index.py).  Using it regularly will
maintain a filesystem-based reflection of the git metadata. This is
good because is means that non-trivial git questions can be answered
with familiar text and file based tools, such as `grep`. The
file-based representation is designed so that ordinary git progress
maps to few file changes.

It would be feasible to maintain this file-based reflection of git
metadata in git. However, that would really only be suitable for going
into a distinct repo. Keeping the file-based reflection of the
metadata of repo X in X itself would create a circular reference
problem: making a new commit in X would entail updating the reflected
metadata which would logically call for another commit.

At present the KubeStellar .gitignore says to ignore the reflection
that is produced by default here.  That means that anyone who cares
can maintain their own local reflection without getting into any
trouble.

## Data model of the file-based reflection of git metadata

The file-based reflection of the git metadata goes into the following
four directories.

- [commits](commits) holds one file for each commit.  The file's name
  is the full hex ID of the commit. The file contains a listing of the
  git ancestors of the file's commit, in commit time order. The line
  for each ancestor holds the ancestor's full hex ID.

- [tags](tags) holds a symlink for each tag.  The symlink's name is
  the sanitized name of the tag, and the link is a relative name for
  the corresponding file in [commits](commits). Sanitization of a name
  replaces special characters with a four-digit hex escape sequence
  "%NNNN".

- [branches](branches) holds a symlink for each local branch.  The
  symlink's name is the sanitized name of the branch, and the link is
  a relative name for the corresponding file in [commits](commits).

- [remotebranches](remotebranches) holds a symlink for each locally
  known branch of a remote.  The name of the symlink is constructed by
  using a dash instead of a slash for the delimiter between the remote
  name and the rest of the reference, and sanitizing the rest of the
  reference.  The value of the link is a relative reference to the
  corresponding commit in [commits](commits).

## Preparing to use the indexing script

The indexing script is written in Python, and it has some
dependencies.  They are documented in
[requirements.txt](scripts/requirements.txt).

I chose to make a Python virtual environment for this thing, in the
usual way.  The [KubeStellar .gitignore](../.gitignore) says to not
keep it in git.

```shell
( cd metadata/scripts
  python3 -m venv venv
  . venv/bin/activate
  pip3 install -r requirements.txt
)
```

Once the virtual environment is created, it can be used at any time in
the usual way.

```shell
cd metadata/scripts
. venv/bin/activate
```

## Usage of the indexing script

With `metadata/scripts` as the current working directory, the
following gets the usage message.

```console
(venv) bash-3.2$ ./index.py -h
usage: index.py [-h] [--repo REPO] [--output OUTPUT]

Parse git metadata to files in index directories commits, tags, branches, remotebranches.

optional arguments:
  -h, --help       show this help message and exit
  --repo REPO      directory holding .git, defaults to ../..
  --output OUTPUT  directory that is parent of index directories, defaults to ..
```

## Example

Ensure the metadata reflection is up-to-date.

With `metadata/scripts` as the current working directory:

```console
(venv) bash-3.2$ ./index.py 

(venv) bash-3.2$ cd ..
```

Following is a brief summary of my current metadata reflection.

```console
(venv) bash-3.2$ ls -l commits | wc
    5027   45236  462405

(venv) bash-3.2$ ls -l tags | wc
      12     123    1228

(venv) bash-3.2$ ls -l branches | wc
      31     332    3497

(venv) bash-3.2$ ls -l remotebranches | wc
     147    1608   17936
```

Following is looking at some commits.

```console
(venv) bash-3.2$ ls commits | head -3
001347ed889fc47d494aafb2bd14a4266edc25c5
001480f6ea4f14bdcbda29982b8c7052fcc158c4
0018008d56afbcfb8817e63c64b120be520989e3

(venv) bash-3.2$ head -3 commits/0018008d56afbcfb8817e63c64b120be520989e3 
85a53bb17697b17174d4861ae81f7b8cbce73798
eab1567199329cb161c8790331992cde7f8b1a3d
afaa573f72905ea047fab6417579672c381336b0
```

Following is looking at some tags.

```console
(venv) bash-3.2$ ls tags | head -3
v0.2.1
v0.2.2
v0.3.0

(venv) bash-3.2$ ls -l tags/v0.3.0
lrwxr-xr-x@ 1 mspreitz  staff  51 Sep 25 22:27 tags/v0.3.0 -> ../commits/bd8088e9d057bc330cce11cac700d7eb9bc44b6e

(venv) bash-3.2$ head -3 tags/v0.3.0
85a53bb17697b17174d4861ae81f7b8cbce73798
eab1567199329cb161c8790331992cde7f8b1a3d
afaa573f72905ea047fab6417579672c381336b0
```

Following is looking at some branches.

```console
(venv) bash-3.2$ ls branches | head -3
alt-return-api
alt-return-impl
cleanup-919a

(venv) bash-3.2$ ls -l branches/cleanup-919a 
lrwxr-xr-x@ 1 mspreitz  staff  51 Sep 25 22:30 branches/cleanup-919a -> ../commits/d15d890bec44b8218e24e6b032b603418bf9cd42

(venv) bash-3.2$ head -3 branches/cleanup-919a 
85a53bb17697b17174d4861ae81f7b8cbce73798
eab1567199329cb161c8790331992cde7f8b1a3d
afaa573f72905ea047fab6417579672c381336b0
```

Following is looking at some remote branches.

```console
(venv) bash-3.2$ ls -l remotebranches/ | head -5
total 0
lrwxr-xr-x@ 1 mspreitz  staff  51 Sep 25 22:30 aa-gh-pages -> ../commits/0a39852ff6e2e83d329a53f3a2bf6fb1133aa0d6
lrwxr-xr-x@ 1 mspreitz  staff  51 Sep 25 22:30 aa-logical-cluster-new-don%0027t-use-to-be-deleted -> ../commits/49a27b8aa3aa5f623575e41831891617253e4cee
lrwxr-xr-x@ 1 mspreitz  staff  51 Sep 25 22:30 aa-logical-clusters -> ../commits/b3a0febe3a04219e1a8613ac46ceba8a52ed27e1
lrwxr-xr-x@ 1 mspreitz  staff  51 Sep 25 22:30 aa-main -> ../commits/75d24fdf26e1c411451eb5477ebf8aa57060b70d

(venv) bash-3.2$ head -3 remotebranches/aa-main
85a53bb17697b17174d4861ae81f7b8cbce73798
eab1567199329cb161c8790331992cde7f8b1a3d
afaa573f72905ea047fab6417579672c381336b0
```

Suppose I found a commit ID somewhere, such as "1a6140f3".  Following
is me looking for what it appears in.

```console
(venv) bash-3.2$ grep ^1a6140f3 tags/*

(venv) bash-3.2$ grep ^1a6140f3 branches/*
branches/main:1a6140f3bc1422ea160479d2043555bd6fd4391d
branches/start-metadata:1a6140f3bc1422ea160479d2043555bd6fd4391d

(venv) bash-3.2$ grep ^1a6140f3 remotebranches/*
remotebranches/origin-HEAD:1a6140f3bc1422ea160479d2043555bd6fd4391d
remotebranches/origin-main:1a6140f3bc1422ea160479d2043555bd6fd4391d
remotebranches/upstream-HEAD:1a6140f3bc1422ea160479d2043555bd6fd4391d
remotebranches/upstream-main:1a6140f3bc1422ea160479d2043555bd6fd4391d
```
