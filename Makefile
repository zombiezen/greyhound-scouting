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
	barcode/barcode.go\
	barcode/code128.go\

CSSFILES=\
    static/css/all.css\
    static/css/generic.css\
    static/css/layout.css\
    static/css/reset.css\
    static/css/style.css\

CLEANFILES=\
	$(TARG)\
	$(CSSFILES)

all: $(TARG) css

clean:
	rm -f $(CLEANFILES)

$(TARG): $(GOFILES)
	go build -o $(TARG)

css: $(CSSFILES)

static/css/%.css: sass/%.scss
	mkdir -p static/css
	sass $< $@

.PHONY: all clean
