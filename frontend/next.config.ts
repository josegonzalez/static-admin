import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  reactStrictMode: true,

  /* enable static export */
  output: "export",

  /* avoid trailing on dynamic routes */
  trailingSlash: false,

  /* disable image optimization since the target is CSR */
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
