# Node.js Matching Engine Dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install --production
COPY . .
EXPOSE 8083
CMD ["node", "matchingEngine.js"]
