import { useCallback, useEffect, useState } from "react";
import { AlertTriangle, Check, CheckCircle2, Loader2, Sparkles, Trash2, X } from "lucide-react";
import type { desktop } from "../../wails/wailsjs/go/models";
import {
  DistillDeleteCandidate,
  DistillEnableCandidate,
  DistillListCandidates,
  DistillRun,
} from "../../wails/wailsjs/go/desktop/App";
import { getSkillCandidateExplanation, t } from "../../lib/i18n";

interface Props {
  open: boolean;
  onClose: () => void;
}

type MessageKind = "info" | "success" | "error";

export function SkillCandidateModal({ open, onClose }: Props) {
  const [candidates, setCandidates] = useState<desktop.SkillCandidateInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [running, setRunning] = useState(false);
  const [busyName, setBusyName] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [messageKind, setMessageKind] = useState<MessageKind>("info");
  const [pendingDelete, setPendingDelete] = useState<desktop.SkillCandidateInfo | null>(null);

  const showMessage = useCallback((kind: MessageKind, text: string) => {
    setMessageKind(kind);
    setMessage(text);
  }, []);

  const loadCandidates = useCallback(async () => {
    setLoading(true);
    try {
      const list = await DistillListCandidates();
      setCandidates(list || []);
    } catch (error) {
      const text = error instanceof Error ? error.message : String(error);
      showMessage("error", text || t("skill_candidate_load_failed"));
    } finally {
      setLoading(false);
    }
  }, [showMessage]);

  useEffect(() => {
    if (open) {
      setMessage(null);
      setMessageKind("info");
      setPendingDelete(null);
      loadCandidates();
    }
  }, [open, loadCandidates]);

  const handleRunDistill = useCallback(async () => {
    setRunning(true);
    setMessage(null);
    try {
      const result = await DistillRun();
      showMessage(result.success ? "success" : "error", result.success ? t("skill_candidate_generated") : t("skill_candidate_generate_failed"));
      await loadCandidates();
    } catch (error) {
      const text = error instanceof Error ? error.message : String(error);
      showMessage("error", text || t("skill_candidate_generate_failed"));
    } finally {
      setRunning(false);
    }
  }, [loadCandidates, showMessage]);

  const handleEnable = useCallback(async (name: string) => {
    setBusyName(name);
    setMessage(null);
    try {
      const result = await DistillEnableCandidate(name);
      showMessage(
        result.success ? "success" : "error",
        result.message || (result.success ? `${t("skill_candidate_added_prefix")}${name}${t("skill_candidate_added_suffix")}` : t("skill_candidate_add_failed")),
      );
      await loadCandidates();
    } catch (error) {
      const text = error instanceof Error ? error.message : String(error);
      showMessage("error", text || t("skill_candidate_add_failed"));
    } finally {
      setBusyName(null);
    }
  }, [loadCandidates, showMessage]);

  const handleDelete = useCallback(async (name: string) => {
    setPendingDelete(null);
    setBusyName(name);
    setMessage(null);
    try {
      const result = await DistillDeleteCandidate(name);
      showMessage(
        result.success ? "success" : "error",
        result.message || (result.success ? `${t("skill_candidate_deleted_prefix")}${name}${t("skill_candidate_deleted_suffix")}` : t("skill_candidate_delete_failed")),
      );
      await loadCandidates();
    } catch (error) {
      const text = error instanceof Error ? error.message : String(error);
      showMessage("error", text || t("skill_candidate_delete_failed"));
    } finally {
      setBusyName(null);
    }
  }, [loadCandidates, showMessage]);

  return (
    <div className={`modal-overlay ${open ? "is-open" : ""}`} onClick={onClose}>
      <div className="modal-dialog relative w-[640px] mx-4" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between px-5 py-4 border-b border-bdr-div">
          <h3 className="text-sm font-medium text-txt flex items-center gap-2">
            <Sparkles className="w-4 h-4 text-accent" />
            {t("skill_candidates")}
          </h3>
          <button onClick={onClose} className="p-1 rounded hover:bg-elevated text-txt-g cursor-pointer" aria-label={t("close")}>
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="p-5 space-y-4 max-h-[70vh] overflow-y-auto">
          <div className="flex items-start justify-between gap-3">
            <p className="text-xs text-txt-g leading-5">
              {t("skill_candidate_description")}
            </p>
            <button
              onClick={handleRunDistill}
              disabled={running || loading}
              className="flex items-center gap-2 px-3 py-1.5 rounded-md bg-accent text-white text-xs hover:bg-accent-light disabled:opacity-60 disabled:cursor-not-allowed cursor-pointer transition-colors"
            >
              {running ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Sparkles className="w-3.5 h-3.5" />}
              {t("skill_candidate_redistill")}
            </button>
          </div>

          {message && (
            <div className={`flex items-start gap-2 rounded-lg border px-3 py-2 text-xs ${
              messageKind === "success"
                ? "border-green-500/25 bg-green-500/10 text-green-600 dark:text-green-400"
                : messageKind === "error"
                  ? "border-red-500/25 bg-red-500/10 text-red-500"
                  : "border-bdr-sub bg-elevated/50 text-txt-m"
            }`}>
              {messageKind === "success" ? <CheckCircle2 className="mt-0.5 h-3.5 w-3.5 flex-shrink-0" /> : null}
              {messageKind === "error" ? <AlertTriangle className="mt-0.5 h-3.5 w-3.5 flex-shrink-0" /> : null}
              <span>{message}</span>
            </div>
          )}

          {loading ? (
            <div className="flex items-center justify-center gap-2 py-12 text-sm text-txt-g">
              <Loader2 className="w-4 h-4 animate-spin" />
              {t("skill_candidate_loading")}
            </div>
          ) : candidates.length === 0 ? (
            <div className="rounded-xl border border-dashed border-bdr-sub p-8 text-center">
              <div className="text-sm text-txt-2 mb-1">{t("skill_candidate_empty_title")}</div>
              <div className="text-xs text-txt-g">{t("skill_candidate_empty_desc")}</div>
            </div>
          ) : (
            <div className="space-y-3">
              {candidates.map((candidate) => {
                const isBusy = busyName === candidate.name;
                return (
                  <div
                    key={candidate.name}
                    className={`rounded-xl border px-4 py-3 transition-colors ${candidate.enabled ? "border-green-500/30 bg-green-500/5" : "border-bdr-sub bg-surface"}`}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <h4 className="text-sm font-medium text-txt truncate">{candidate.name}</h4>
                          {candidate.enabled && (
                            <span className="inline-flex items-center gap-1 rounded-full bg-green-500/10 px-2 py-0.5 text-[10px] font-medium text-green-500">
                              <CheckCircle2 className="h-3 w-3" />
                              {t("skill_candidate_added_badge")}
                            </span>
                          )}
                        </div>
                        <p className="mt-1 text-xs text-txt-g leading-5">{candidate.explanation || getSkillCandidateExplanation(candidate) || candidate.description || t("skill_candidate_no_description")}</p>
                        {candidate.enabled && (
                          <p className="mt-1 text-[11px] text-green-600 dark:text-green-400">
                            {t("skill_candidate_enabled_list")}
                          </p>
                        )}
                      </div>
                      <div className="text-xs text-txt-m flex-shrink-0">
                        {t("skill_candidate_confidence") + " "}{Math.round((candidate.confidence || 0) * 100)}%
                      </div>
                    </div>

                    {candidate.pattern && (
                      <div className="mt-3 rounded-md bg-elevated/60 px-2.5 py-2 font-mono text-[11px] text-txt-m break-all">
                        {candidate.pattern}
                      </div>
                    )}

                    <div className="mt-3 flex items-center justify-between gap-2">
                      <span className="text-[11px] text-txt-g">
                        {candidate.enabled ? t("skill_candidate_status_added") : t("skill_candidate_status_pending")}
                      </span>
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => setPendingDelete(candidate)}
                          disabled={isBusy}
                          className="flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-xs text-red-400 hover:bg-red-500/10 disabled:opacity-60 disabled:cursor-not-allowed cursor-pointer transition-colors"
                        >
                          {isBusy ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Trash2 className="w-3.5 h-3.5" />}
                          {t("skill_candidate_delete_action")}
                        </button>
                        <button
                          onClick={() => handleEnable(candidate.name)}
                          disabled={candidate.enabled || isBusy}
                          className="flex items-center gap-1.5 rounded-md bg-accent px-2.5 py-1.5 text-xs text-white hover:bg-accent-light disabled:opacity-60 disabled:cursor-not-allowed cursor-pointer transition-colors"
                        >
                          {isBusy ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Check className="w-3.5 h-3.5" />}
                          {candidate.enabled ? t("skill_candidate_added_badge") : t("skill_candidate_add_action")}
                        </button>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {pendingDelete && (
          <div className="absolute inset-0 z-20 flex items-center justify-center bg-black/30 px-4" onClick={() => setPendingDelete(null)}>
            <div className="w-full max-w-sm rounded-xl border border-bdr bg-surface p-4 shadow-xl" onClick={(e) => e.stopPropagation()}>
              <div className="flex items-start gap-3">
                <div className="rounded-full bg-red-500/10 p-2 text-red-500">
                  <AlertTriangle className="h-4 w-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <h4 className="text-sm font-medium text-txt">{t("skill_candidate_confirm_delete_title")}</h4>
                  <p className="mt-1 text-xs leading-5 text-txt-g">
                    {t("skill_candidate_confirm_delete_prefix")}<span className="font-mono text-txt-m">{pendingDelete.name}</span>{t("skill_candidate_confirm_delete_suffix")}
                  </p>
                </div>
              </div>
              <div className="mt-4 flex justify-end gap-2">
                <button
                  onClick={() => setPendingDelete(null)}
                  className="rounded-md bg-elevated px-3 py-1.5 text-xs text-txt-2 hover:bg-elevated/80 cursor-pointer"
                >
                  {t("cancel")}
                </button>
                <button
                  onClick={() => handleDelete(pendingDelete.name)}
                  className="rounded-md bg-red-500 px-3 py-1.5 text-xs text-white hover:bg-red-600 cursor-pointer"
                >
                  {t("skill_candidate_confirm_delete_action")}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
