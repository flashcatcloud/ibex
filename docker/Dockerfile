FROM ubuntu:21.04

WORKDIR /app
ADD ibex /app
RUN chmod +x ibex

EXPOSE 10090
EXPOSE 20090

CMD ["/app/ibex", "-h"]