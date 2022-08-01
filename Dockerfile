FROM scratch
COPY lucky /
EXPOSE 16601
WORKDIR /goodluck
ENTRYPOINT ["/lucky"]
CMD ["-c", "/goodluck/lucky.conf"]