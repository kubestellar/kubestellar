# How to Sign Your Pull Requests

In order to get your pull requests approved, you must first complete a DCO sign-off. To do this, you must have a GPG and SSH key. Basic setup instructions can be found below (For more detailed instructions, refer to the Github [GPG](https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key) and [SSH](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent#generating-a-new-ssh-key) setup pages):

---

Before starting, make sure that your user email is verified on Github. Git Bash is also highly recommended. 

<br />

## Setting up the GPG Key

1. Install [GnuPG](https://www.gnupg.org/download/) (the GPG command line tool).
   - Binary releases for your specific OS can be found after scrolling down to the Binary Releases section (i.e. Gpg4win on Windows, Mac GPG for macOS, etc).
   - Follow the prompts on the installer to set up GnuPG.

2. Open Git Bash and use the following command to generate your GPG key pair: 
   ```shell
   gpg --full-generate-key
   ```
3. If prompted to specify the size, type, and duration of the key that you want, press ```Enter``` to select the default option.
4. Once prompted, enter your user info and a passphrase:
   - Make sure to list your email as the same one that's verified by Github
5. Use the following command to list the long form of your generated GPG keys:
   ```shell
   gpg --list-secret-keys --keyid-format=long
   ```
   - Your GPG key ID should be the characters on the output line starting with ```sec```, beginning directly after the ```/``` and ending before the listed date.
     - For example, in the output below (from the Github [GPG](https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key) setup page), the GPG key ID would be ```3AA5C34371567BD2```
     ```shell
     $ gpg --list-secret-keys --keyid-format=long
      /Users/hubot/.gnupg/secring.gpg
      ------------------------------------
      sec   4096R/3AA5C34371567BD2 2016-03-10 [expires: 2017-03-10]
      uid                          Hubot <hubot@example.com>
      ssb   4096R/4BB6D45482678BE3 2016-03-10

     ```
6. Copy your GPG key ID and run the command below, replacing ```[your_GPG_key_ID]``` with the key ID you just copied:
   ```shell
   gpg --armor --export [your_GPG_key_ID]
   ```
7. This should generate an output with your GPG key. Copy the characters starting from ```-----BEGIN PGP PUBLIC KEY BLOCK-----``` and ending at ```--END PGP PUBLIC KEY BLOCK-----``` (inclusive) to your clipboard.
8. After copying or saving your GPG key, navigate to **Settings** in your Github
9. Navigate to the **SSH and GPG keys** page under the Access section in the sidebar
10. Under GPG keys, select **New GPG key**
    - Enter a suitable name for your key under "Title" and paste your GPG key that you copied/saved in **Step 7** under "Key".
    - Once done, click **Add GPG key**
11. Your new GPG key should now be displayed under GPG keys.

<br />

## Setting up the SSH Key

1. Open Git Bash and use the following command to generate your new SSH key (make sure to replace ```your_email``` with your Github-verified email address):
   ```shell
   ssh-keygen -t ed25519 -C "your_email"
   ```
   
2. Press ```Enter``` to select the default option if prompted to set a save-file or passphrase for the key (you may choose to enter a passphrase if desired; this will prompt you to enter the passphrase everytime you perform a DCO sign-off).
   - The following output should generate a randomart image 
3. Use the following command to copy the new SSH key to your clipboard:
   ```shell
   clip < ~/.ssh/id_ed25519.pub
   ```
4. After copying or saving your SSH key, navigate to **Settings** in your Github.
5. Navigate to the **SSH and GPG keys** page under the Access section in the sidebar.
6. Under SSH keys, select **New SSH key**.
   - Enter a suitable name for your key under "Title"
   - Open the dropdown menu under "Key type" and select **Signing Key**
   - Paste your SSH key that you copied/saved in **Step 3** under "Key"
7. Your new SSH key should now be displayed under SSH keys.

<br />

## Creating Pull Requests

Whether it's editing files from Kubestellar.io or directly from the Kubestellar Github, there are a couple steps to follow that streamlines the workflow of your PR:

1. Changes made to any file are automatically committed to a new branch in your fork.
   - When committing, make sure to specify the type of PR at the beginning of your commit message (i.e. :bug: if it addresses a bug-type issue)
   - If the PR addresses a specific issue that has already been opened in the github, make sure to include the opened issue in **additional comments** (i.e. "fixes Issue #2XXX")
     
2. Click **Propose Changes** after writing the commit message, review your changes, and then create the PR.
3. If your PR addresses an already opened issue on the github, make sure to close the issue once your PR is approved and closed.

<br />

## Signing Pull Requests

Your submitted PR must pass the automated checks in order to be reviewed. This requires for you to perform a DCO sign-off for your PR. The following instructions provide a basic walkthrough if you have already set up your GPG and SSH keys:

1. Navigate to the **Code** page of the Kubestellar github.
   
2. Click the **Fork** dropdown in the top right corner of the page.
   - Under "Existing Forks" click your fork (should look something like "your_username/kubestellar")
3. Once in your fork, click the **Code** dropdown.
   - Under the "Local" tab at the top of the dropdown, select the SSH tab
   - Copy the SSH repo URL to your clipboard
4. Open Git Bash (create or change to a different directory if desired).
5. Clone the repository using ```git clone``` followed by pasting the URL you just copied.
6. Change your directory to the Kubestellar repo using ```cd kubestellar```.
7. ```git checkout``` to the branch in your fork where the changes were committed.
   - The branch name should be written at the top of your submitted PR page and looks something like "patch-*X*" (where "X" should be the number of PRs made on your fork to date)
8. Once in your branch, type ```git commit -s --amend``` to sign off your PR.
   - You may replace ```--amend``` with a ```-m``` followed by a commit message if you desire; the ```--amend``` simply uses the same commit message as the one you wrote when initially submitting the PR
   - If prompted with a sign-off page in your Git Bash, type ```:wq!``` to exit the prompt
9. Type ```git push -f origin [branch_name]```, replacing ```[branch_name]``` with the actual name of your branch.
10. Navigate back to your PR github page.
    - A green ```dco-signoff: yes``` label indicates that your PR is successfully signed









