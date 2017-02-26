BINARY=sgo

build:
	if [ ! -d ${GOPATH}/src/github.com/PuerkitoBio/goquery	  ] ; then go get -x github.com/PuerkitoBio/goquery	 ; fi
	if [ ! -d ${GOPATH}/src/github.com/cavaliercoder/grab	  ] ; then go get -x github.com/cavaliercoder/grab	 ; fi
	if [ ! -d ${GOPATH}/src/github.com/gizak/termui		  ] ; then go get -x github.com/gizak/termui		 ; fi
	if [ ! -d ${GOPATH}/src/github.com/mholt/archiver	  ] ; then go get -x github.com/mholt/archiver	         ; fi
	if [ ! -d ${GOPATH}/src/github.com/olekukonko/tablewriter ] ; then go get -x github.com/olekukonko/tablewriter   ; fi
	go build -o ${BINARY}

install: build
	go install -x ${BINARY}.go

clean:
	if [ -f *~ ] ; then rm *~ ; fi
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	if [ -f ${GOPATH}/bin/${BINARY} ] ; then rm ${GOPATH}/bin/${BINARY} ; fi
	if [ -d ${GOPATH}/src/github.com/PuerkitoBio/goquery	  ] ; then rm -rf ${GOPATH}/src/github.com/PuerkitoBio/goquery   	 ; fi
	if [ -d ${GOPATH}/src/github.com/cavaliercoder/grab	  ] ; then rm -rf ${GOPATH}/src/github.com/cavaliercoder/grab    	 ; fi
	if [ -d ${GOPATH}/src/github.com/gizak/termui		  ] ; then rm -rf ${GOPATH}/src/github.com/gizak/termui	       	 ; fi
	if [ -d ${GOPATH}/src/github.com/mholt/archiver	          ] ; then rm -rf ${GOPATH}/src/github.com/mholt/archiver	         ; fi
	if [ -d ${GOPATH}/src/github.com/olekukonko/tablewriter   ] ; then rm -rf ${GOPATH}/src/github.com/olekukonko/tablewriter       ; fi

.PHONY: clean install