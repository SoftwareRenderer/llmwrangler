# LLaMA Wrangler (a llama.cpp multiplexer)
<p align="center">
  <img width="256" height="256" src="https://github.com/SoftwareRenderer/llmwrangler/assets/138734813/7ca3b4a3-809b-44e9-94be-2dff5892019c">
</p>

This improves hardware utilization for projects that implement [llama.cpp](https://github.com/ggerganov/llama.cpp/). The idea is that workload should be assigned to the fastest machine until response times get slow enough to assign work to the rest of the workers.

This only works for the plain llama.cpp `/completion` endpoint.

TODO: Automate management of hosts

## How to use (Docker)
1. Update docker-compose with your preferred LISTEN port (default: 7000), and optionally a llama.cpp host
2. Start the container
3. Navigate to http://localhost:7000 to add/remove hosts
