FROM eclipse-temurin:17-jdk-alpine
WORKDIR /app
COPY ../../../../../pom.xml ./
COPY ../../../../../src ./src
RUN ./mvnw dependency:resolve
RUN ./mvnw package -DskipTests
EXPOSE 8082
CMD ["java", "-cp", "target/classes:target/dependency/*", "com.botlogictrade.MatchingEngineRedisService"]
