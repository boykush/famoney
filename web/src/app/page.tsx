import { checkHealth } from "./lib/api";

export const dynamic = "force-dynamic";

export default async function Home() {
	const results = await Promise.all([
		checkHealth("Expense Service", "/api/v1/expenses/health"),
	]);

	return (
		<div className="py-10">
			<div className="text-center mb-10">
				<h1 className="text-3xl font-bold mb-2">Famoney</h1>
				<p className="text-gray-600">Service Health Status</p>
			</div>

			<div className="grid gap-4 max-w-2xl mx-auto">
				{results.map((result) => (
					<div
						key={result.service}
						className="bg-white rounded-lg border border-gray-200 p-5 flex items-center justify-between"
					>
						<div>
							<h2 className="font-semibold text-lg">{result.service}</h2>
							<p className="text-sm text-gray-500 font-mono">
								{result.endpoint}
							</p>
						</div>
						<span
							className={`px-3 py-1 rounded-full text-sm font-medium ${
								result.ok
									? "bg-green-100 text-green-800"
									: "bg-red-100 text-red-800"
							}`}
						>
							{result.status}
						</span>
					</div>
				))}
			</div>
		</div>
	);
}
