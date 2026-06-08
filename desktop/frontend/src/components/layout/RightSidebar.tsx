import {
  Wrench,
  CheckCircle,
  Loader2,
  XCircle,
  FileText,
  ChevronDown,
  ChevronRight,
  ListChecks,
  X,
} from "lucide-react";
import { useState } from "react";
import { useActivityStore, type ActivityEntry } from "../../stores/activityStore";
import { t } from "../../lib/i18n";

function ActivityEntryItem({ entry }: { entry: ActivityEntry }) {
  const [expanded, setExpanded] = useState(false);

  const statusIcon =
    entry.status === "running" ? (
      <Loader2 className="w-3.5 h-3.5 text-amber-400 animate-spin" />
    ) : entry.status === "done" ? (
      <CheckCircle className="w-3.5 h-3.5 text-green-500" />
    ) : (
      <XCircle className="w-3.5 h-3.5 text-red-500" />
    );

  return (
    <div className="border-b border-bdr-sub/50 last:border-b-0">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex items-center gap-2 w-full px-3 py-2 text-xs hover:bg-elevated/30 transition-colors cursor-pointer"
      >
        {expanded ? (
          <ChevronDown className="w-3 h-3 text-txt-g flex-shrink-0" />
        ) : (
          <ChevronRight className="w-3 h-3 text-txt-g flex-shrink-0" />
        )}
        {statusIcon}
        <span className="font-mono text-txt-2 truncate">{entry.name}</span>
        <span className="ml-auto text-txt-m text-[10px] flex-shrink-0">
          {new Date(entry.timestamp).toLocaleTimeString(undefined, {
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
          })}
        </span>
      </button>
      {expanded && entry.detail && (
        <pre className="px-5 pb-2 text-[11px] text-txt-g whitespace-pre-wrap break-all max-h-40 overflow-y-auto">
          {entry.detail.length > 500
            ? entry.detail.slice(0, 500) + "..."
            : entry.detail}
        </pre>
      )}
    </div>
  );
}

export function RightSidebar() {
  const entries = useActivityStore((s) => s.entries);
  const plan = useActivityStore((s) => s.plan);
  const fileDiffs = useActivityStore((s) => s.fileDiffs);
  const setRightSidebarOpen = useActivityStore((s) => s.setRightSidebarOpen);

  return (
    <div className="flex flex-col h-full w-[320px]">
      {/* Header with close button */}
      <div className="px-3 py-2.5 border-b border-bdr-sub flex items-center gap-2">
        <Wrench className="w-4 h-4 text-txt-g" />
        <span className="text-sm font-medium text-txt">{t("activity")}</span>
        <span className="text-[10px] text-txt-m ml-auto mr-2">
          {entries.length}
        </span>
        <button
          onClick={() => setRightSidebarOpen(false)}
          className="p-1 rounded hover:bg-elevated text-txt-g hover:text-txt transition-colors cursor-pointer"
          title={`${t("close")} (Ctrl+I)`}
        >
          <X className="w-3.5 h-3.5" />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto">
        {/* Plan Panel */}
        {plan && (
          <div className="border-b border-bdr-sub p-3">
            <div className="flex items-center gap-2 mb-2">
              <ListChecks className="w-3.5 h-3.5 text-accent" />
              <span className="text-xs font-medium text-txt-2">
                {plan.goal}
              </span>
            </div>
            <div className="w-full h-1.5 bg-elevated rounded-full mb-2 overflow-hidden">
              <div
                className="h-full bg-accent rounded-full transition-all duration-300"
                style={{
                  width: `${
                    plan.totalSteps > 0
                      ? (plan.steps.filter((s) => s.status === "completed")
                          .length /
                          plan.totalSteps) *
                        100
                      : 0
                  }%`,
                }}
              />
            </div>
            <div className="space-y-1">
              {plan.steps.map((step) => (
                <div
                  key={step.id}
                  className="flex items-center gap-2 text-xs py-0.5"
                >
                  {step.status === "completed" ? (
                    <CheckCircle className="w-3 h-3 text-green-500 flex-shrink-0" />
                  ) : step.status === "in_progress" ? (
                    <Loader2 className="w-3 h-3 text-amber-400 animate-spin flex-shrink-0" />
                  ) : step.status === "failed" ? (
                    <XCircle className="w-3 h-3 text-red-500 flex-shrink-0" />
                  ) : (
                    <span className="w-3 h-3 rounded-full border border-txt-g flex-shrink-0" />
                  )}
                  <span
                    className={`${
                      step.status === "in_progress"
                        ? "text-txt"
                        : step.status === "completed"
                        ? "text-txt-m"
                        : "text-txt-g"
                    }`}
                  >
                    {step.description}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* File Changes */}
        {fileDiffs.length > 0 && (
          <div className="border-b border-bdr-sub p-3">
            <div className="flex items-center gap-2 mb-2">
              <FileText className="w-3.5 h-3.5 text-blue-400" />
              <span className="text-xs font-medium text-txt-2">
                {t("file_changes")}
              </span>
            </div>
            <div className="space-y-1">
              {fileDiffs.map((diff) => (
                <div
                  key={diff.path}
                  className="flex items-center gap-2 text-xs font-mono"
                >
                  <span className="text-txt-2 truncate flex-1">
                    {diff.path}
                  </span>
                  <span className="text-green-500 text-[10px]">
                    +{diff.additions}
                  </span>
                  <span className="text-red-400 text-[10px]">
                    -{diff.deletions}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Activity Log */}
        <div>
          {entries.length === 0 && (
            <div className="px-4 py-8 text-center text-txt-m text-xs">
              {t("no_activity")}
            </div>
          )}
          {entries.map((entry) => (
            <ActivityEntryItem key={entry.id} entry={entry} />
          ))}
        </div>
      </div>
    </div>
  );
}
