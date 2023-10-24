### Testing a KubeStellar documentation PR

Here are the steps to checkout a git pull request for local testing.

**STEP-1: Checkout the Pull Request**

Helpers: [GitHub](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/checking-out-pull-requests-locally), [DevOpsCube](https://devopscube.com/checkout-git-pull-request/)

#### 1.1 Use the pull request number to fetch origin (note: be sure to check out the right branch!)

&nbsp;&nbsp;&nbsp;&nbsp;Fetch the reference to the pull request based on its ID number, creating a new branch locally. Replace ID with your PR # and BRANCH_NAME with the desired branch name.

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;```git fetch origin pull/ID/head:BRANCH_NAME```  

#### 1.2 Switch to the new branch

&nbsp;&nbsp;&nbsp;Checkout the BRANCH_NAME where you have all the changes from the pull request.

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;```git switch BRANCH_NAME```  

&nbsp;&nbsp;&nbsp;&nbsp;At this point, you can do anything you want with this branch. You can run some local tests, or merge other branches into the branch.

**STEP-2: Test and Build the Documentation (optional)**

&nbsp;&nbsp;&nbsp;&nbsp;Use this procedure if you want to view and modify the documentation in the branch you have checked out.

Helpers: [KubeStellar/docs](https://github.com/kubestellar/kubestellar/tree/main/docs), [MkDocs](https://www.mkdocs.org/user-guide/installation/)

#### 2.1 Install MkDocs and its requirements

```
  cd docs
  pip install mkdocs
  pip install -r requirements.txt  
```

#### 2.2 Build and view the documentation

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;```mkdocs serve```

&nbsp;&nbsp;&nbsp;&nbsp;Next, open a browser to [http://127.0.0.1:8000](http://127.0.0.1:8000) and review the changes.