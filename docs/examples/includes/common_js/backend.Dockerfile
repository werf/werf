FROM node:18-alpine
WORKDIR /app
COPY backend/ /app/
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]
