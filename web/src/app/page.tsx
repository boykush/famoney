"use client";

import { useEffect, useState } from "react";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export default function Home() {
	const [authenticated, setAuthenticated] = useState(false);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		const token = localStorage.getItem("access_token");
		if (!token) {
			window.location.href = `${API_BASE_URL}/auth/login`;
			return;
		}

		fetch(`${API_BASE_URL}/auth/me`, {
			headers: { Authorization: `Bearer ${token}` },
		})
			.then((res) => {
				if (res.status === 401) {
					localStorage.removeItem("access_token");
					window.location.href = `${API_BASE_URL}/auth/login`;
					return;
				}
				if (!res.ok) {
					throw new Error(`HTTP ${res.status}`);
				}
				setAuthenticated(true);
			})
			.catch((e) => {
				setError(e instanceof Error ? e.message : "Unknown error");
			})
			.finally(() => {
				setLoading(false);
			});
	}, []);

	if (loading) {
		return (
			<div className="py-10 text-center">
				<p className="text-gray-500">Loading...</p>
			</div>
		);
	}

	if (error) {
		return (
			<div className="py-10 text-center">
				<p className="text-red-600">Error: {error}</p>
			</div>
		);
	}

	if (authenticated) {
		return (
			<div className="py-10">
				<div className="text-center mb-10">
					<h1 className="text-3xl font-bold mb-2">Famoney</h1>
					<p className="text-gray-600">Authenticated</p>
				</div>
			</div>
		);
	}

	return null;
}
