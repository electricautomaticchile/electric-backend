import http from "k6/http";
import { check } from "k6";
import exec from "k6/execution";

const BASE_URL = (__ENV.BASE_URL || "http://localhost:4000").replace(/\/+$/, "");
const START_RATE = Number(__ENV.START_RATE || 10);
const PEAK_RATE = Number(__ENV.PEAK_RATE || 1000);
const PRE_ALLOCATED_VUS = Number(__ENV.PRE_ALLOCATED_VUS || 300);
const MAX_VUS = Number(__ENV.MAX_VUS || 2000);
const IP_SPREAD = __ENV.IP_SPREAD !== "false";
const RAMP_25_DURATION = __ENV.RAMP_25_DURATION || "2m";
const RAMP_50_DURATION = __ENV.RAMP_50_DURATION || "3m";
const RAMP_PEAK_DURATION = __ENV.RAMP_PEAK_DURATION || "5m";
const HOLD_DURATION = __ENV.HOLD_DURATION || "5m";
const RAMP_DOWN_DURATION = __ENV.RAMP_DOWN_DURATION || "1m";

export const options = {
  scenarios: {
    landing_leads: {
      executor: "ramping-arrival-rate",
      startRate: START_RATE,
      timeUnit: "1s",
      preAllocatedVUs: PRE_ALLOCATED_VUS,
      maxVUs: MAX_VUS,
      stages: [
        { target: Math.round(PEAK_RATE * 0.25), duration: RAMP_25_DURATION },
        { target: Math.round(PEAK_RATE * 0.5), duration: RAMP_50_DURATION },
        { target: PEAK_RATE, duration: RAMP_PEAK_DURATION },
        { target: PEAK_RATE, duration: HOLD_DURATION },
        { target: 0, duration: RAMP_DOWN_DURATION },
      ],
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    "http_req_duration{endpoint:lead_submit}": ["p(95)<500", "p(99)<1500"],
  },
};

function virtualIp(iteration) {
  const a = Math.floor(iteration / 65536) % 250;
  const b = Math.floor(iteration / 256) % 256;
  const c = iteration % 256;
  return `10.${a}.${b}.${c}`;
}

export default function () {
  const iteration = exec.scenario.iterationInTest;
  const isInvestor = iteration % 2 === 0;
  const id = `${iteration}-${Date.now()}`;
  const payload = {
    type: isInvestor ? "investor" : "distributor",
    name: `Lead ${id}`,
    email: `lead-${id}@example.com`,
    organization: isInvestor ? "Acme Ventures" : "Distribuidora Norte",
    message: "Prueba de carga k6",
    extra: isInvestor
      ? { ticket: "USD 10K - 25K" }
      : { role: "Operaciones", meterCount: "10.000 - 100.000" },
    turnstileToken: __ENV.TURNSTILE_TOKEN || "",
  };

  const headers = {
    "Content-Type": "application/json",
    Accept: "application/json",
  };
  if (IP_SPREAD) headers["X-Forwarded-For"] = virtualIp(iteration);

  const response = http.post(`${BASE_URL}/api/leads`, JSON.stringify(payload), {
    headers,
    tags: { endpoint: "lead_submit" },
  });

  check(response, {
    "lead created": (r) => r.status === 201,
  });
}
