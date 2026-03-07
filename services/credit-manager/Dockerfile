FROM --platform=$BUILDPLATFORM node:18.19.0 AS builder

WORKDIR /workshop
COPY package.json package-lock.json ./
RUN npm config set registry https://registry.npmmirror.com/
RUN npm install --frozen-lockfile
COPY . .
RUN NODE_OPTIONS="--max-old-space-size=4096" npm run build
FROM --platform=$BUILDPLATFORM nginx:stable-alpine AS runner

RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d/

COPY --from=builder /workshop/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
