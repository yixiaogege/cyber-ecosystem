import fc from "fast-check";
import { describe, expect, it } from "vitest";

import { fixed, fmt, niceStepMax, pct, signedPP } from "./format";

describe("fmt", () => {
  it("groups thousands and keeps integers whole", () => {
    expect(fmt(0)).toBe("0");
    expect(fmt(999)).toBe("999");
    expect(fmt(1234)).toBe("1,234");
    expect(fmt(1234567)).toBe("1,234,567");
    expect(fmt(1234567n)).toBe("1,234,567");
  });
});

describe("pct / fixed", () => {
  it("renders with the requested digits", () => {
    expect(pct(12.345)).toBe("12.3%");
    expect(pct(12.345, 2)).toBe("12.35%");
    expect(fixed(3.14159, 2)).toBe("3.14");
    expect(fixed(3.14159)).toBe("3.1");
  });
});

describe("signedPP", () => {
  it("signs positive deltas explicitly, never zero", () => {
    expect(signedPP(2.5)).toBe("+2.5pp");
    expect(signedPP(-2.5)).toBe("-2.5pp");
    expect(signedPP(0)).toBe("0.0pp");
  });
});

describe("niceStepMax", () => {
  it("degrades non-positive inputs to the 0..4 ladder floor", () => {
    expect(niceStepMax(0)).toBe(4);
    expect(niceStepMax(-3)).toBe(4);
  });

  it("picks ladder steps from the table", () => {
    expect(niceStepMax(7)).toBe(8); // raw 1.75 → step 2
    expect(niceStepMax(9)).toBe(10); // raw 2.25 → step 2.5
    expect(niceStepMax(1_000_000)).toBe(1_000_000); // raw 250k → step 2.5×1e5
  });

  it("keeps the ceiling integral, covering, and at most 1.5× the data max", () => {
    fc.assert(
      fc.property(fc.double({ min: 4, max: 1e12, noNaN: true }), (value) => {
        const max = niceStepMax(value);
        expect(Number.isInteger(max)).toBe(true);
        expect(max).toBeGreaterThanOrEqual(value);
        expect(max).toBeLessThanOrEqual(value * 1.5 + 1e-9);
      }),
    );
  });
});
