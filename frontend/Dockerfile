# Use Node
FROM node:20

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy rest
COPY . .

# Build
RUN npm run build

# Expose port
EXPOSE 3000

# Start
CMD ["npm", "run", "start"]