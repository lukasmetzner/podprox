FROM debian
COPY ./bin/podprox /podprox
ENTRYPOINT /podprox