### Testing a KubeStellar documentation PR

- checkout the documentation related PR locally on your system. note: be sure to check out the right branch!  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;e.g. ```git clone -b user-quickstart https://github.com/clubanderson/kubestellar.git```
- install mkdocs ([https://www.mkdocs.org/user-guide/installation/](https://www.mkdocs.org/user-guide/installation/))
- cd to /docs and run ```pip install -r requirements.txt```
- after all requirements are installed, run ```mkdocs serve```
- open browser to [http://127.0.0.1:8000](http://127.0.0.1:8000)