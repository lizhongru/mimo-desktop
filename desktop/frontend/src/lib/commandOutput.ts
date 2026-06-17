export function stripAnsi(input: string): string {
  return input
    .replace(/[\u001B\u009B][[\]()#;?]*(?:(?:(?:[a-zA-Z\d]*(?:;[a-zA-Z\d]*)*)?\u0007)|(?:(?:\d{1,4}(?:;\d{0,4})*)?[\dA-PR-TZcf-nq-uy=><~]))/g, "")
    .replace(/\r\n/g, "\n")
    .replace(/\r/g, "\n");
}

export function summarizeCommandOutput(output?: string): string {
  const clean = stripAnsi(output || "")
    .split("\n")
    .map((line) => line.trimEnd())
    .filter((line) => line.trim().length > 0);

  if (clean.length === 0) return "";

  const important = clean.filter((line) => {
    const lower = line.toLowerCase();
    return (
      lower.includes("warning") ||
      lower.includes("error") ||
      lower.includes("failed") ||
      lower.includes("built in") ||
      lower.includes("✓ built") ||
      lower.includes("chunks are larger") ||
      lower.includes("to load an es module")
    );
  });

  const selected = important.length > 0 ? important : clean.slice(-6);
  return selected.slice(-8).join("\n").trim();
}
