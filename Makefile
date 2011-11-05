include $(GOROOT)/src/Make.inc

GCIMPORTS=-I$(GOPATH)/pkg/$(GOOS)_$(GOARCH)
LDIMPORTS=-L$(GOPATH)/pkg/$(GOOS)_$(GOARCH)

TARG=scouting
GOFILES=\
	main.go\
	model.go\
	team.go\

include $(GOROOT)/src/Make.cmd
