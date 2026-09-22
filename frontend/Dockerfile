FROM node:22-alpine

WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN printf 'packages: []\nallowBuilds:\n  esbuild: true\n  lightningcss: true\n  unrs-resolver: true\n  msw: false\n' > pnpm-workspace.yaml \
    && corepack enable \
    && pnpm install --frozen-lockfile

COPY . ./

ARG VITE_API_BASE_URL=http://localhost:8080
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN pnpm exec vite build

EXPOSE 3000

CMD ["pnpm", "exec", "vite", "preview", "--host", "0.0.0.0", "--port", "3000"]
