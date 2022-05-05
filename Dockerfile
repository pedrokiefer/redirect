FROM scratch
COPY redirect /bin/
COPY ./ui /ui
EXPOSE 10100
EXPOSE 10101
VOLUME /etc/redirect

CMD ["/bin/redirect", "-config", "/etc/redirect/config.json", "-bind", "0.0.0.0:10100"]
