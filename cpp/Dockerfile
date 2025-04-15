FROM gcc:13
WORKDIR /app
COPY . .
RUN apt-get update && apt-get install -y cmake libssl-dev && rm -rf /var/lib/apt/lists/*
RUN cmake . && make
EXPOSE 8085
CMD ["/app/matching_engine"]
