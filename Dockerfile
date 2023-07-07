FROM golang:1.17

# download the pim, build it and install it
RUN go get github.com/dsblox/pim/... # change 18

# set up some aliases useful in our development environment
RUN echo 'alias cd-pim="cd /go/pkg/mod/github.com/dsblox/pim@v0.1.1"' >> ~/.bashrc
RUN echo 'alias run-pim="cd-pim;pim -server -db yaml"' >> ~/.bashrc
RUN echo 'alias make-pim="cd-pim;go install"' >> ~/.bashrc

# whatever environment we use must make sure enough source code is mounted at /app to run the server
WORKDIR /go/pkg/mod/github.com/dsblox/pim@v0.1.1

# CMD is only executed if another command is not specified on the docker run command
# so if container is run as a daemon then assume we are running the server
# but if container is run with -it and /bin/bash as the command then the server won't be started
# . and we can build and restart the server in a dev / test environment.
CMD pim -server

# when we have a server we can uncomment any exposed port
# we want ports to be more flexible and not be hard-coded into our image
# so I'm commenting this out and will modify the code to read the port
# from an environment variable and only default to 4000 if it can't find it.
EXPOSE 4000
