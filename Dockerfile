FROM golang:1.18-alpine
WORKDIR /app
COPY . .
#COPY [^.]* .
RUN go build -o app .
RUN find . -type f -name "*.go" -exec rm -fv {} \;
RUN find . -type f -name '.*' -exec rm -fv {} \;
RUN rm -rvf controllers globals helpers middleware routes .git
RUN find . -type f -name '.code-workspace' -exec rm -fv {} \;
EXPOSE 8080
CMD ["./app"]
