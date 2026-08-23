FROM mcr.microsoft.com/playwright:v1.49.1-jammy
WORKDIR /tests
COPY package.json ./
RUN npm install
COPY . .
ENV PYTHONUNBUFFERED=1
RUN apt-get update && apt-get install -y python3 && rm -rf /var/lib/apt/lists/*
