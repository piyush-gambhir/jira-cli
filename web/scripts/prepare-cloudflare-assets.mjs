import { cp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";

const sitePath = process.argv[2];

if (!sitePath || !/^[a-z0-9-]+$/.test(sitePath)) {
  throw new Error("Expected a URL-safe site path argument.");
}

const source = path.resolve("out");
const destinationRoot = path.resolve(".cloudflare/assets");
const destination = path.join(destinationRoot, sitePath);

await rm(destinationRoot, { recursive: true, force: true });
await mkdir(destination, { recursive: true });
await cp(source, destination, { recursive: true });

// Wrangler only reads _redirects from the assets root, and the site is served
// under /<sitePath>, so move the file up and prefix every path rule with it.
const siteRedirects = path.join(destination, "_redirects");
const rules = await readFile(siteRedirects, "utf8").catch((error) => {
  if (error.code === "ENOENT") return null;
  throw error;
});

if (rules !== null) {
  const prefix = (value) => (value.startsWith("/") ? `/${sitePath}${value}` : value);
  const rewritten = rules
    .split("\n")
    .map((line) => {
      const trimmed = line.trim();
      if (trimmed === "" || trimmed.startsWith("#")) return line;
      const [from, to, ...rest] = trimmed.split(/\s+/);
      if (!to) return line;
      return [prefix(from), prefix(to), ...rest].join(" ");
    })
    .join("\n");
  await writeFile(path.join(destinationRoot, "_redirects"), rewritten);
  await rm(siteRedirects);
}
