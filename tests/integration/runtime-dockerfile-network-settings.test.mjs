import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import test from "node:test";

const root = resolve(".");

test("runtime dependency stages configure resilient pnpm networking", async () => {
  const recipe = await readFile(resolve(root, "deploy/runtime/Dockerfile"), "utf8");
  const stages = new Map(
    recipe.split(/^(?=FROM )/mu).map((stage) => [stage.match(/^FROM \S+ AS (\S+)$/mu)?.[1], stage]),
  );

  const settings = [
    "PNPM_CONFIG_NETWORK_CONCURRENCY=4",
    "PNPM_CONFIG_FETCH_RETRIES=5",
    "PNPM_CONFIG_FETCH_RETRY_MINTIMEOUT=10000",
    "PNPM_CONFIG_FETCH_RETRY_MAXTIMEOUT=60000",
    "PNPM_CONFIG_FETCH_TIMEOUT=300000",
  ];

  const dependencyStages = ["openclaw-dependency-inputs"];
  for (const name of dependencyStages) {
    const stage = stages.get(name);
    assert.ok(stage, `runtime Dockerfile must define ${name}`);
    for (const setting of settings) {
      assert.ok(stage.includes(`ENV ${setting}`), `${name} must configure ${setting}`);
    }
  }

  for (const setting of settings) {
    assert.equal(
      recipe.split(`ENV ${setting}`).length - 1,
      dependencyStages.length,
      `${setting} must appear only in the configured parent dependency stages`,
    );
  }
});
