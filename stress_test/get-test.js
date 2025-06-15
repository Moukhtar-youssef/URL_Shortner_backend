import http from "k6/http";
import { check, group } from "k6";
import { Trend, Rate } from "k6/metrics";

// Custom metrics
let responseTimes = new Trend("response_times");
let errorRate = new Rate("errors");

export let options = {
  stages: [
    { duration: "30s", target: 100 }, // Gradual ramp-up
    { duration: "1m", target: 500 }, // Intermediate load
    { duration: "2m", target: 1000 }, // Peak load
    { duration: "30s", target: 0 }, // Ramp-down
  ],
  thresholds: {
    "http_req_duration{status:200}": ["p(95)<100"], // Only successful requests
    http_req_failed: ["rate<0.005"], // Stricter error threshold
  },
};

export default function () {
  group("Backend response test", function () {
    let params = {
      redirects: 0,
      timeout: "5s",
      tags: { name: "ShortURLRedirect" },
    };

    let res = http.get("http://localhost:8080/TByPyLV", params);

    // Record metrics
    responseTimes.add(res.timings.duration);
    errorRate.add(res.status >= 400);

    // Validate response
    check(res, {
      "correct redirect status": (r) => r.status === 301 || r.status === 302,
      "has location header": (r) => r.headers["Location"] !== undefined,
    });
  });
}
