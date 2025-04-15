FROM rust:1.76
WORKDIR /app
COPY . .
RUN cargo build --release
EXPOSE 8084
CMD ["/app/target/release/matching_engine"]
