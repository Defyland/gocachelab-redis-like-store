import http from "k6/http";
import { check } from "k6";

export const options = {
  vus: 1,
  iterations: 5,
};

export default function () {
  const baseURL = __ENV.BASE_URL || "http://127.0.0.1:8080";
  const health = http.get(`${baseURL}/healthz`);
  check(health, {
    "health status is 200": (res) => res.status === 200,
  });

  const metrics = http.get(`${baseURL}/metrics`);
  check(metrics, {
    "metrics status is 200": (res) => res.status === 200,
    "metrics include key gauge": (res) => res.body.includes("gocachelab_keys"),
  });
}

