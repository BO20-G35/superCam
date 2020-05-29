FROM ubuntu:latest
RUN apt-get update && apt-get install -y supervisor

WORKDIR /bin

COPY supercam.exe .
COPY supercam_init /etc/init.d/supercam_init
COPY *.html /bin/
ADD assets /bin/assets

EXPOSE 8082

CMD ["bash", "/etc/init.d/supercam_init"]