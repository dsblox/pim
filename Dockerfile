FROM golang:1.8

# download the pim, build it and install it
RUN go get github.com/dsblox/pim/... #force by changing this number 8

# set up some aliases useful in our development environment
RUN echo 'alias cd-pim="cd /go/src/github.com/dsblox/pim"' >> ~/.bashrc
RUN echo 'alias run-pim="cd-pim;pim"' >> ~/.bashrc
RUN echo 'alias make-pim="cd-pim;go install"' >> ~/.bashrc

# CMD is only executed if another command is not specified on the docker run command
# so if container is run as a daemon then assume we are running the server
# but if container is run with -it and /bin/bash as the command then the server won't be started
# . and we can build and restart the server in a dev / test environment.
CMD pim -server

# when we have a server we can uncomment any exposed port
EXPOSE 4000
