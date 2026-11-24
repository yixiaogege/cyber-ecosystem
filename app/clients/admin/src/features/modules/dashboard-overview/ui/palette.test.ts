import { describe, expect, it } from "vitest";

import { AXIS_TEXT, GRID, ORDINAL_FUNNEL, ORDINAL_LEVEL, S1, S2 } from "./palette";

const HEX = /^#[0-9a-f]{6}$/;

describe("series colors", () => {
  it("keeps the main and comparison series distinguishable", () => {
    expect(S1).not.toBe(S2);
  });

  it("anchors the main series to the funnel's primary ordinal", () => {
    // The main sequence (new registrations) and the funnel's first-strong
    // stage share one hue anchor: cross-chart comparison stays stable.
    expect(S1).toBe(ORDINAL_FUNNEL[1]);
  });
});

describe("ordinal ramps", () => {
  it("sizes the ramps to their stages: five funnel steps, four fission levels", () => {
    expect(ORDINAL_FUNNEL).toHaveLength(5);
    expect(ORDINAL_LEVEL).toHaveLength(4);
  });

  it("uses distinct valid hex entries only", () => {
    for (const ramp of [ORDINAL_FUNNEL, ORDINAL_LEVEL]) {
      for (const color of ramp) expect(color).toMatch(HEX);
      expect(new Set(ramp).size).toBe(ramp.length);
    }
  });
});

describe("chart skeleton", () => {
  it("stays neutral: grid and axis text are fixed grays, not theme tokens", () => {
    expect(GRID).toMatch(HEX);
    expect(AXIS_TEXT).toMatch(HEX);
  });
});
