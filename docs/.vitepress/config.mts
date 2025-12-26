import { defineConfig } from "vitepress";

const docsRoot = "https://docs.lightz.exchange";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Lightz Client",
  description: "Lightz Client Docs",
  head: [["link", { rel: "icon", href: "/assets/logo.svg" }]],
  themeConfig: {
    logo: "/assets/logo.svg",
    search: {
      provider: "local",
      options: {
        detailedView: true
      }
    },
    nav: [{ text: "🏠 Docs Home", link: docsRoot, target: "_self" }],
    sidebar: [
      {
        items: [
          { text: "👋 Introduction", link: "/index" },
          { text: "💰 Wallets", link: "/wallets" },
          { text: "🔁 Autoswap", link: "/autoswap" },
          { text: "🏅 Lightz Pro", link: "/lightzd-pro" },
          { text: "🎛️ Configuration", link: "/configuration" },
          { text: "🤖 gRPC API", link: "/grpc" },
          {
            text: "🤖 REST API",
            link: "https://github.com/lightzapp/lightzd-client/blob/master/pkg/lightzrpc/rest-annotations.yaml",
          },
          { text: "🏠 Docs Home", link: docsRoot, target: "_self" },
        ],
      },
    ],
    socialLinks: [
      {
        icon: "github",
        link: "https://github.com/lightzapp/lightzd-client",
      },
    ],
  },
  // Ignore dead links to localhost
  ignoreDeadLinks: [/https?:\/\/localhost/],
});
