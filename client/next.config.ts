import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  experimental: {
    turbo: {
      rules: {
        '*.phaser.js': {
          loaders: ['ignore-loader'],
        },
      },
    },
  },
  webpack: (config, { isServer }) => {
    // Phaser는 클라이언트 사이드에서만 실행
    if (isServer) {
      config.resolve.alias = {
        ...config.resolve.alias,
        phaser: false,
      };
    }
    
    return config;
  },
};

export default nextConfig;
