const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface HealthStatus {
	service: string;
	endpoint: string;
	status: string;
	ok: boolean;
}

export async function checkHealth(
	service: string,
	path: string,
): Promise<HealthStatus> {
	try {
		const res = await fetch(`${API_BASE_URL}${path}`, { cache: "no-store" });
		if (!res.ok) {
			return {
				service,
				endpoint: path,
				status: `HTTP ${res.status}`,
				ok: false,
			};
		}
		const data = await res.json();
		return {
			service,
			endpoint: path,
			status: data.status || "UNKNOWN",
			ok: data.status === "SERVING",
		};
	} catch (e) {
		return {
			service,
			endpoint: path,
			status: e instanceof Error ? e.message : "ERROR",
			ok: false,
		};
	}
}
