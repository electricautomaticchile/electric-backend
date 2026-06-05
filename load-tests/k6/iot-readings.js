import http from "k6/http";
import { check } from "k6";
import exec from "k6/execution";

const BASE_URL = (__ENV.BASE_URL || "http://localhost:4000").replace(/\/+$/, "");
const IOT_API_KEY = __ENV.IOT_API_KEY || "";
const START_RATE = Number(__ENV.START_RATE || 100);
const PEAK_RATE = Number(__ENV.PEAK_RATE || 3000);
const PRE_ALLOCATED_VUS = Number(__ENV.PRE_ALLOCATED_VUS || 500);
const MAX_VUS = Number(__ENV.MAX_VUS || 5000);
const RAMP_25_DURATION = __ENV.RAMP_25_DURATION || "2m";
const RAMP_50_DURATION = __ENV.RAMP_50_DURATION || "3m";
const RAMP_PEAK_DURATION = __ENV.RAMP_PEAK_DURATION || "5m";
const HOLD_DURATION = __ENV.HOLD_DURATION || "10m";
const RAMP_DOWN_DURATION = __ENV.RAMP_DOWN_DURATION || "1m";
const IP_SPREAD = __ENV.IP_SPREAD !== "false";
const DEVICE_IDS = (__ENV.DEVICE_IDS || "")
  .split(",")
  .map((id) => id.trim())
  .filter(Boolean);

export const options = {
  scenarios: {
    iot_readings: {
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
    "http_req_duration{endpoint:iot_reading}": ["p(95)<250", "p(99)<750"],
  },
};

function deviceId(iteration) {
  if (DEVICE_IDS.length > 0) return DEVICE_IDS[iteration % DEVICE_IDS.length];
  return `MED-${String(iteration % 10000).padStart(6, "0")}`;
}

function virtualIp(iteration) {
  const a = Math.floor(iteration / 65536) % 250;
  const b = Math.floor(iteration / 256) % 256;
  const c = iteration % 256;
  return `10.${a}.${b}.${c}`;
}

export default function () {
  const iteration = exec.scenario.iterationInTest;
  const payload = {
    deviceId: deviceId(iteration),
    voltaje: 220 + (iteration % 10),
    corriente: 4 + (iteration % 8) / 10,
    potencia: 900 + (iteration % 100),
    energia: 1000 + iteration / 100,
    frecuencia: 50,
    factorPotencia: 0.96,
    timestamp: Date.now(),
  };

  const headers = {
    "Content-Type": "application/json",
    Accept: "application/json",
    "X-IoT-API-Key": IOT_API_KEY,
  };
  if (IP_SPREAD) headers["X-Forwarded-For"] = virtualIp(iteration);

  const response = http.post(`${BASE_URL}/api/iot/lectura`, JSON.stringify(payload), {
    headers,
    tags: { endpoint: "iot_reading" },
  });

  check(response, {
    "reading accepted": (r) => r.status === 200,
  });
}
