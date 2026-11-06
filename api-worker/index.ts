import { DurableObject } from "cloudflare:workers";

// Cloudflare front-door Worker for the Treewiki Go gateway.
// All requests route to ONE named container ("singleton") because the
// gateway runs in-memory: separate instances would hold divergent data.

const PORT = 8080;
// Keep the container warm 30 minutes: in-memory wiki data survives while warm.
const INACTIVITY_TIMEOUT_MS = 30 * 60 * 1000;
const READY_ATTEMPTS = 120;

interface Env {
	GATEWAY_CONTAINER: DurableObjectNamespace;
	JWT_SECRET: string;
	CORS_ALLOWED_ORIGINS: string;
}

export class GatewayContainer extends DurableObject<Env> {
	private starting: Promise<void> | undefined;

	async fetch(request: Request): Promise<Response> {
		this.starting ??= this.startAndWaitForPort().finally(() => {
			this.starting = undefined;
		});
		await this.starting;

		const url = new URL(request.url);
		url.protocol = "http:";
		url.host = "container";
		const forwarded = new Request(url, request);
		forwarded.headers.delete("host");
		return this.ctx.container!.getTcpPort(PORT).fetch(forwarded);
	}

	private async startAndWaitForPort(): Promise<void> {
		const container = this.ctx.container!;
		if (!container.running) {
			container.start({
				image: container.images.base,
				env: {
					PORT: String(PORT),
					ENV: "production",
					JWT_SECRET: this.env.JWT_SECRET,
					CORS_ALLOWED_ORIGINS: this.env.CORS_ALLOWED_ORIGINS,
				},
			});
		}
		await container.setInactivityTimeout(INACTIVITY_TIMEOUT_MS);

		// running means startup was requested, not that the HTTP server is ready.
		const port = container.getTcpPort(PORT);
		for (let attempt = 0; attempt < READY_ATTEMPTS; attempt++) {
			try {
				const response = await port.fetch("http://container/healthz", {
					signal: AbortSignal.timeout(1000),
				});
				await response.body?.cancel();
				if (response.ok) return;
			} catch {
				// The port may not be listening yet. Retry the readiness probe only.
			}
			await scheduler.wait(1000);
		}
		throw new Error("Gateway container did not become ready on port 8080");
	}
}

export default {
	async fetch(request: Request, env: Env): Promise<Response> {
		try {
			return await env.GATEWAY_CONTAINER.getByName("singleton").fetch(request);
		} catch (error) {
			console.error("Gateway container request failed:", error);
			return new Response("Gateway container unavailable. Retry shortly.", {
				status: 502,
			});
		}
	},
};
