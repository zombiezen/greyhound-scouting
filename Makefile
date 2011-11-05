include $(GOROOT)/src/Make.inc

GCIMPORTS=-I$(GOPATH)/pkg/$(GOOS)_$(GOARCH)
LDIMPORTS=-L$(GOPATH)/pkg/$(GOOS)_$(GOARCH)

TARG=scouting
GOFILES=\
	main.go\
	model.go\
	server.go\
	team.go\

include $(GOROOT)/src/Make.cmd

CSSFILES=\
    static/css/all.css\
    static/css/generic.css\
    static/css/layout.css\
    static/css/reset.css\
    static/css/style.css\

CLEANFILES+=$(CSSFILES)

static/css/%.css: sass/%.scss
	mkdir -p static/css
	sass $< $@

all: css
css: $(CSSFILES)
