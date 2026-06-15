import { useState } from "react";
import { ChevronRight, ChevronDown } from "lucide-react";

interface Props {
  data: unknown;
  level?: number;
}

export function JsonTree({ data, level = 0 }: Props) {
  if (data === null) {
    return <span className="text-txt-m italic font-mono text-[11px] opacity-60">null</span>;
  }
  if (data === undefined) {
    return <span className="text-txt-m italic font-mono text-[11px] opacity-60">undefined</span>;
  }
  if (typeof data === "string") {
    return (
      <span className="text-emerald-500 dark:text-emerald-400 font-mono text-[11px] break-all leading-relaxed">
        &quot;{data.length > 500 ? data.slice(0, 500) + "..." : data}&quot;
      </span>
    );
  }
  if (typeof data === "number") {
    return <span className="text-orange-500 dark:text-orange-300 font-mono text-[11px]">{data}</span>;
  }
  if (typeof data === "boolean") {
    return <span className="text-purple-500 dark:text-purple-400 font-mono text-[11px]">{String(data)}</span>;
  }
  if (Array.isArray(data)) {
    return <JsonArray data={data} level={level} />;
  }
  return <JsonObject data={data as Record<string, unknown>} level={level} />;
}

// ── Row: fixed left gutter for toggle + right content ──

function JsonRow({
  expanded,
  onToggle,
  hasChildren,
  children,
}: {
  expanded?: boolean;
  onToggle?: () => void;
  hasChildren?: boolean;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-start">
      {/* Left gutter — fixed width, only for toggle buttons */}
      <div className="w-5 flex-shrink-0 flex items-center justify-center pt-[1px]">
        {hasChildren && onToggle && (
          <button
            onClick={onToggle}
            className="w-4 h-4 flex items-center justify-center rounded hover:bg-elevated/50 cursor-pointer transition-colors"
          >
            {expanded ? (
              <ChevronDown className="w-3 h-3 text-txt-m opacity-60" />
            ) : (
              <ChevronRight className="w-3 h-3 text-txt-m opacity-60" />
            )}
          </button>
        )}
      </div>
      {/* Right content */}
      <div className="min-w-0 flex-1">
        {children}
      </div>
    </div>
  );
}

// ── Object ─────────────────────────────────────────────

function JsonObject({ data, level }: { data: Record<string, unknown>; level: number }) {
  const [expanded, setExpanded] = useState(level < 3);
  const keys = Object.keys(data);

  if (keys.length === 0) {
    return <span className="text-txt-m font-mono text-[11px]">{"{}"}</span>;
  }

  if (!expanded) {
    return (
      <JsonRow expanded={false} onToggle={() => setExpanded(true)} hasChildren>
        <span className="font-mono text-[11px] leading-[1.7]">
          <span className="text-txt-m/70">{"{ "}</span>
          <span className="text-txt-m/50">{keys.length} {keys.length === 1 ? "key" : "keys"}</span>
          <span className="text-txt-m/70">{" }"}</span>
        </span>
      </JsonRow>
    );
  }

  return (
    <div className="font-mono text-[11px] leading-[1.7]">
      <JsonRow expanded onToggle={() => setExpanded(false)} hasChildren>
        <span className="text-txt-m/70">{"{"}</span>
      </JsonRow>
      <div className="ml-5 border-l border-bdr-sub/50 pl-2.5 my-0.5">
        {keys.map((key) => (
          <div key={key} className="py-[1px] flex items-baseline gap-0">
            <span className="text-blue-500 dark:text-sky-300 flex-shrink-0">&quot;{key}&quot;</span>
            <span className="text-txt-m/40 mx-1.5 flex-shrink-0">:</span>
            <span className="min-w-0 flex-1">
              <JsonTree data={data[key]} level={level + 1} />
            </span>
          </div>
        ))}
      </div>
      <JsonRow hasChildren={false}>
        <span className="text-txt-m/70">{"}"}</span>
      </JsonRow>
    </div>
  );
}

// ── Array ──────────────────────────────────────────────

function JsonArray({ data, level }: { data: unknown[]; level: number }) {
  const [expanded, setExpanded] = useState(level < 3);

  if (data.length === 0) {
    return <span className="text-txt-m font-mono text-[11px]">[]</span>;
  }

  if (!expanded) {
    return (
      <JsonRow expanded={false} onToggle={() => setExpanded(true)} hasChildren>
        <span className="font-mono text-[11px] leading-[1.7]">
          <span className="text-txt-m/70">{"[ "}</span>
          <span className="text-txt-m/50">{data.length} {data.length === 1 ? "item" : "items"}</span>
          <span className="text-txt-m/70">{" ]"}</span>
        </span>
      </JsonRow>
    );
  }

  return (
    <div className="font-mono text-[11px] leading-[1.7]">
      <JsonRow expanded onToggle={() => setExpanded(false)} hasChildren>
        <span className="text-txt-m/70">[</span>
      </JsonRow>
      <div className="ml-5 border-l border-bdr-sub/50 pl-2.5 my-0.5">
        {data.map((item, idx) => (
          <div key={idx} className="py-[1px] flex items-baseline gap-0">
            <span className="text-txt-m/40 flex-shrink-0 w-6 text-right mr-1.5 tabular-nums font-mono">{idx}</span>
            <span className="min-w-0 flex-1">
              <JsonTree data={item} level={level + 1} />
            </span>
          </div>
        ))}
      </div>
      <JsonRow hasChildren={false}>
        <span className="text-txt-m/70">]</span>
      </JsonRow>
    </div>
  );
}
