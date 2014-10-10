.PHONY: test
test:
	pylint taskcluster
	./bin/python setup.py test

JS_CLIENT_BRANCH=master
APIS_JSON=$(PWD)/apis.json
APIS_JS_HREF=https://raw.githubusercontent.com/taskcluster/taskcluster-client/master/lib/apis.js

.PHONY: apis.json
apis.json:
	@echo Downloading $(APIS_JS_HREF)
	curl -L -o apis.js $(APIS_JS_HREF)
	OUTPUT=$(APIS_JSON) node translateApis.js
	@python -mjson.tool $(APIS_JSON) > /dev/null || echo "apis.json cannot be parsed by python's JSON"
	
	
	

.PHONY: dev-env
dev-env:
	virtualenv .
	./bin/pip install -r requirements.txt
	./bin/pip install -r devel-requirements.txt
