import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  turbopack: {
    root: __dirname,
  },
  async redirects() {
    return [
      { source: "/dizimo", destination: "/doacoes", permanent: true },
      { source: "/dizimo/:path*", destination: "/doacoes/:path*", permanent: true },
    ];
  },
};

export default nextConfig;
