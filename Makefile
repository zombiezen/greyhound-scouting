include $(GOROOT)/src/Make.inc

TARG=scouting
GOFILES=\
	event.go\
	main.go\
	model.go\
	paging.go\
	reports.go\
	server.go\
	store.go\
	tags.go\
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
