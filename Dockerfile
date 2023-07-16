FROM golang:1.17

# create an app directory where the running (production) environment will live
# note that we can mount dev environments as well, which will be in a
# separate directory and will contain all needed source code.  /app is
# designed to hold only the code needed at runtime (but for now has additional code)
WORKDIR /app

# get our go.mod from local file system so we can install our code
COPY go.mod go.sum ./

# get our code's dependencies based on instructions in go.mod
RUN go mod download

# now copy our code from local host into the container
COPY . .

# build the code
RUN go install

# set up some aliases useful in our development environment
RUN echo 'alias cd-pim="cd /app"' >> ~/.bashrc # this may be wrong
RUN echo 'alias run-pim="cd-pim;pim -server -db yaml"' >> ~/.bashrc
RUN echo 'alias make-pim="cd-pim;go install"' >> ~/.bashrc

# CMD is only executed if another command is not specified on the docker run command
# so if container is run as a daemon then assume we are running the server
# but if container is run with -it and /bin/bash as the command then the server won't be started
CMD pim -server

# when we have a server we can uncomment any exposed port
# we want ports to be more flexible and not be hard-coded into our image
# so I'm commenting this out and will modify the code to read the port
# from an environment variable and only default to 4000 if it can't find it.
EXPOSE 4000
