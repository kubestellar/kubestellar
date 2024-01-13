#!/usr/bin/env python3

import argparse
import html.parser
import os
import sys

def alist_get(alist, key):
    for assoc in alist:
        if key == assoc[0]:
            return assoc[1]
    return None

class div_bits():
    cls = None
    code = 0
    pre = 0
    anchor_id = None
    def __init__(self, cls):
        self.cls = cls
        return
    def __str__(self) -> str:
        return f'(cls={self.cls}, code={self.code}, pre={self.pre}, anchor_id={self.anchor_id})'
    def should_extract(self) -> bool:
        return self.code > 0 and self.pre > 0 and  self.cls in ('language-shell highlight')
    pass

class MyHTMLParser(html.parser.HTMLParser):
    divs = []
    last_collected_anchor_id = None

    def __init__(self, outfile):
        super().__init__(convert_charrefs=True)
        self.outfile = outfile

    def handle_starttag(self, tag, attrs):
        if tag == "pre":
            if self.divs:
                self.divs[0].pre += 1
        elif tag == "code":
            if self.divs:
                self.divs[0].code += 1
        elif tag == "a":
            id = alist_get(attrs, "id")
            if  self.divs:
                self.divs[0].anchor_id = id
        elif tag == "div":
            this_class = alist_get(attrs, "class")
            db = div_bits(this_class)
            self.divs = [db] + self.divs
        return
    
    def handle_endtag(self, tag):
        if tag == "pre":
            if self.divs:
                self.divs[0].pre -= 1
        elif tag == "code":
            if self.divs:
                self.divs[0].code = 1
        elif tag == "div":
            if self.divs:
                self.divs = self.divs[1:]
        return
    
    def handle_data(self, data:str):
        if self.divs :
            if self.divs[0].should_extract():
                if self.last_collected_anchor_id and self.last_collected_anchor_id != self.divs[0].anchor_id:
                    print(f'Imagining newline at {self.getpos()}, new anchor={self.divs[0].anchor_id}', file=sys.stderr)
                    # self.outfile.write("\n")
                print(f'Adding data (len={len(data)}) at {self.getpos()}', file=sys.stderr)
                self.outfile.write(data)
                self.last_collected_anchor_id = self.divs[0].anchor_id
            else:
                print(f'Skipping data (len={len(data)}) at {self.getpos()} because current div is {self.divs[0]}', file=sys.stderr)
        return
    pass

def extract(infile, outfile):
    parser = MyHTMLParser(outfile)
    indata = infile.read()
    parser.feed(indata)
    parser.close()
    return

def cli():
    argp = argparse.ArgumentParser(add_help=True,
                                   description='extract bash snippets from HTML')
    argp.add_argument('infile')
    argp.add_argument('outfile')
    args = argp.parse_args()
    infile_name = args.infile
    outfile_name = args.outfile
    with open(infile_name) as infile:
        with open(outfile_name, mode='wt') as outfile:
            extract(infile, outfile)
    return

if __name__ == "__main__":
    cli()
