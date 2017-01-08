# Use the golang image
FROM golang
MAINTAINER landon.wainwright@gmail.com

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/landonia/gocollect

# Create the directory for the user.db
RUN mkdir -p /usr/local/gocollect

# Go and fetch all the dependencies
WORKDIR /go/src/github.com/landonia/gocollect
RUN go get

# Install the program
RUN go install github.com/landonia/gocollect

# Run the gocollect program by default when the container starts.
ENTRYPOINT ["/go/bin/gocollect", "-db=/usr/local/gocollect/user.db", "-addr=:8090", "-loglevel=info"]

# Document that the service listens on port 8090.
EXPOSE 8090
