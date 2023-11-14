{%
   include-markdown "../../common-subs/coming-soon.md"
   start="<!--coming-soon-start-->"
   end="<!--coming-soon-end-->"
%}

<!-- Code management
  Prow, Gh actions broken links, pr verifier, emoji in titles of prs, add issue to project. Add pr to project. Check spelling errors, wordlist.txt, 
Quay.io -->
# Code Management

## Fork kubestellar into your own repo, create a local branch, set upstream to kubestellar, add and commit changes to local branch, and squash your commits

### Fork the Github kubestellar repo into your own Github repo:
You can do this either 1: from the kubestellar Github website using the "Fork" button or 2: by using the git fork command from your local git command line interface, such as git bash.

copy the forked repo from Github to your local system by using the "git clone" command or by downloading the repository's zip file.

In your new local forked repo, set upstream to kubestellar main

check what your repo's remote settings are
```
git remote -v
```

### Set upstream to use kubestellar: 
```
git remote add upstream git@github.com:kubestellar/kubestellar.git
```

For example:
```
owner@BOOK-U0EMIUAFHD MINGW64 ~/src/edge-mc (main)
 git remote -v
origin  git@github.com:fileppb/edge-mc.git (fetch)
origin  git@github.com:fileppb/edge-mc.git (push)

owner@BOOK-U0EMIUAFHD MINGW64 ~/src/edge-mc (main)
 git remote add upstream git@github.com:kubestellar/kubestellar.git

owner@BOOK-U0EMIUAFHD MINGW64 ~/src/edge-mc (main)
 git remote -v
origin  git@github.com:fileppb/edge-mc.git (fetch)
origin  git@github.com:fileppb/edge-mc.git (push)
upstream        git@github.com:kubestellar/kubestellar.git (fetch)
upstream        git@github.com:kubestellar/kubestellar.git (push)

owner@BOOK-U0EMIUAFHD MINGW64 ~/src/edge-mc (main)
 git fetch upstream
Enter passphrase for key '/c/Users/owner/.ssh/id_rsa':
remote: Enumerating objects: 60394, done.
remote: Counting objects: 100% (5568/5568), done.
remote: Compressing objects: 100% (255/255), done.
remote: Total 60394 (delta 4768), reused 5457 (delta 4706), pack-reused 54826
Receiving objects: 100% (60394/60394), 52.38 MiB | 3.25 MiB/s, done.
Resolving deltas: 100% (34496/34496), completed with 415 local objects.

owner@BOOK-U0EMIUAFHD MINGW64 ~/src/edge-mc (main)
```
 git status

On branch main
Your branch is up to date with 'origin/main'.

nothing to commit, working tree clean
```
```
### Select an issue to work on and create a local branch, 

Create a local branch for your work, preferably including the issue number in the branch name

for example if working on issue #11187, then you might name your local branch "issue-1187"
```
git checkout -b issue-1187
```

### As you work and change files, you should try to commit relatively small pieces of work, using the following commands
```
git add (there are several options you can specify for the git add command)

git commit -m "your message"

git push -u origin branch-name (-u sets upstream to origin which is your remote github repository)
```
### When you have completed your work and tested it locally, then you should perform a squash of the git commits to make the upcoming push request more manageable.

To perform a squash, checkout the branch you want to squash,
1. use the "git log" command to see the history of commits to the branch
2. Count the number of commits you want to squash
3. use the "git rebase -i HEAD~n" where n is the number of commits you would like to squash together. (There are other ways to do this)
4. The text editor you have configured to use with git should automagically open your source and you will see a list of commits preceded by "pick". Leaving the first "pick" as it is, replace the remaining "pick"s with "squash"es. 
5. Save the text file and exit the editor.
6. The text editor will open again to let you edit comments for your new squashed commit.
7. Make your edits if any and save and exit the file.
The commits will then be squashed into one commit.
8. run the "git push --force-with-lease" command

### When you are done, push your changes to your remote branch
```
git push -u origin <branch-name>
```

## Create a Pull Request (PR) from your Github repo branch in order to request review and approval from the Kubestellar team

Take a look at https://github.com/kubestellar/kubestellar/blob/main/CONTRIBUTING.md

You can create a Pull Request from your Github web repository by selecting the "Compare & pull request" button.

You be presented with a Github web page titled Comparing Changes, which allows you enter metadata regarding your pull request

Reference the issue you are addressing ( add #issue-number)
Add one of the listed emojis to the first character of the title of your new PR indicating the type of issue (bug fix, feature, etc)
Complete the summary description field
Complete the Related issue field by inserting the issue number preceded by the # character, for example "#1187"
Decide whether this is a draft PR or if it's ready for review, and select the option you want by expanding on the Create Pull Reuest button.
Assign a label to the PR from the available list of labels (a drop down list on the right side of the web page)

Kubestellar CI pipeline:

Prow (https://docs.prow.k8s.io/docs/overview/)
