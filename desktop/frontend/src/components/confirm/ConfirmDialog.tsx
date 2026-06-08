import { AlertTriangle, CheckCircle, XCircle, Shield } from "lucide-react";
import type { ConfirmAction } from "../../lib/types";
import { t } from "../../lib/i18n";

interface Props {
  action: ConfirmAction;
  onApprove: () => void;
  onDeny: () => void;
  onApproveAll: () => void;
}

function levelColor(level: string): string {
  switch (level?.toLowerCase()) {
    case "critical":
      return "text-red-500 bg-red-500/10 border-red-500/30";
    case "high":
      return "text-orange-400 bg-orange-500/10 border-orange-500/30";
    case "medium":
      return "text-yellow-400 bg-yellow-500/10 border-yellow-500/30";
    case "low":
      return "text-green-400 bg-green-500/10 border-green-500/30";
    default:
      return "text-txt-g bg-elevated border-bdr";
  }
}

export function ConfirmDialog({ action, onApprove, onDeny, onApproveAll }: Props) {
  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 backdrop-blur-sm">
      <div className="bg-surface border border-bdr rounded-xl w-full max-w-md mx-4 shadow-2xl">
        {/* Header */}
        <div className="flex items-center gap-3 px-5 py-4 border-b border-bdr-sub">
          <Shield className="w-5 h-5 text-amber-400" />
          <h2 className="text-base font-medium text-txt">
            {t("safety_confirm_title")}
          </h2>
        </div>

        {/* Content */}
        <div className="px-5 py-4 space-y-3">
          <div className="flex items-center gap-2">
            <span
              className={`text-xs font-medium px-2 py-0.5 rounded border ${levelColor(
                action.level
              )}`}
            >
              {action.level?.toUpperCase() || t("unknown")}
            </span>
            <span className="text-sm text-txt-2 font-mono">
              {action.tool}
            </span>
          </div>

          <p className="text-sm text-txt-g">{action.description}</p>

          {action.params && Object.keys(action.params).length > 0 && (
            <div className="bg-elevated/60 rounded-lg p-3 text-xs font-mono space-y-1">
              {Object.entries(action.params).map(([key, value]) => (
                <div key={key} className="flex gap-2">
                  <span className="text-txt-m">{key}:</span>
                  <span className="text-txt-2 break-all">
                    {typeof value === "string"
                      ? value.length > 200
                        ? value.slice(0, 200) + "..."
                        : value
                      : JSON.stringify(value)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 px-5 py-4 border-t border-bdr-sub">
          <button
            onClick={onDeny}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm bg-elevated text-txt-2 hover:bg-elevated/80 transition-colors"
          >
            <XCircle className="w-3.5 h-3.5" />
            {t("deny")}
          </button>
          <button
            onClick={onApproveAll}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm bg-elevated text-txt-2 hover:bg-elevated/80 transition-colors"
          >
            <CheckCircle className="w-3.5 h-3.5" />
            {t("approve_all")}
          </button>
          <button
            onClick={onApprove}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm bg-accent/20 text-accent hover:bg-accent/30 transition-colors ml-auto"
          >
            <AlertTriangle className="w-3.5 h-3.5" />
            {t("approve")}
          </button>
        </div>
      </div>
    </div>
  );
}
