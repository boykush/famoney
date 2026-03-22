"use client";

import { useSearchParams } from "next/navigation";
import { Suspense, useEffect } from "react";

function CallbackHandler() {
	const searchParams = useSearchParams();

	useEffect(() => {
		const token = searchParams.get("access_token");
		if (token) {
			localStorage.setItem("access_token", token);
			window.location.href = "/";
		}
	}, [searchParams]);

	return (
		<div className="py-10 text-center">
			<p className="text-gray-500">Authenticating...</p>
		</div>
	);
}

export default function AuthCallback() {
	return (
		<Suspense
			fallback={
				<div className="py-10 text-center">
					<p className="text-gray-500">Loading...</p>
				</div>
			}
		>
			<CallbackHandler />
		</Suspense>
	);
}
