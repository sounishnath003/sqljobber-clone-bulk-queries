import http from "k6/http";
import { group, sleep, check } from "k6";

export const options = {
  // A number specifying the number of VUs to run concurrently.
  vus: 10,
  // A string specifying the total duration of the test run.
  duration: "10s",
};

// The function that defines VU logic.
//
// See https://grafana.com/docs/k6/latest/examples/get-started-with-k6/ to learn more
// about authoring k6 scripts.
//

const API_URL = "http://localhost:3000";

export default function () {
  group("/healthy", function () {
    const response = http.get(`${API_URL}/healthy`);
    check(response, {
      "response code was 200?": (res) => res.status === 200,
      "response json as status 200?": (res) =>
        JSON.parse(res.body)["status"] == 200,
    });
    sleep(1);
  });

  group("/api/v1/tasks", function () {
    let response = http.get(`${API_URL}/api/v1/tasks`);
    check(response, {
      "/api/v1/tasks response code was 200?": (res) => res.status === 200,
      "/api/v1/tasks response json as status 200?": (res) =>
        JSON.parse(res.body)["status"] == 200,
    });
    check(response, {
      "/api/v1/tasks response code was 200?": (res) => res.status === 200,
      "/api/v1/tasks response json as status 200?": (res) =>
        JSON.parse(res.body)["status"] == 200,
    });

    const jobID = `job_${Date.now()}`;

    response = http.post(
      `${API_URL}/api/v1/tasks/slow_query/jobs`,
      JSON.stringify({
        job_id: jobID,
        args: ["10"],
        db: "postgres",
      })
    );

    check(response, {
      "/api/v1/tasks/slow_query/jobs response code was 200?": (res) =>
        res.status === 200,
      "/api/v1/tasks/slow_query/jobs response json as status 200?": (res) =>
        JSON.parse(res.body)["status"] == 200,
    });

    response = http.get(`${API_URL}/api/v1/jobs/${jobID}`);
    check(response, {
      "/api/v1/jobs/jobid response code was 200?`": (res) =>
        res.status === 200,
      "/api/v1/jobs/jobid response json as status 200?": (res) =>
        JSON.parse(res.body)["status"] == 200,
      "/api/v1/jobs/jobid job has error?": (res) =>
        JSON.parse(res.body)["data"]["error"] == "",
    });
  });

  sleep(3);
}
