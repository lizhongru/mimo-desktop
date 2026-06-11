import { useEffect, useRef, useState } from "react";
import { Brain, Check, ChevronDown, ChevronRight, Cpu } from "lucide-react";
import { useSettingsStore } from "../../stores/settingsStore";
import { t } from "../../lib/i18n";
import { useAnimatedOpen } from "../../lib/useAnimatedOpen";

export function ModelReasoningPicker() {
  const [rawOpen, setRawOpen] = useState(false);
  const { shouldRender: open, closing } = useAnimatedOpen(rawOpen, 150);
  const currentModel = useSettingsStore((s) => s.currentModel);
  const currentModelKey = useSettingsStore((s) => s.currentModelKey);
  const models = useSettingsStore((s) => s.models);
  const reasoningLevel = useSettingsStore((s) => s.reasoningLevel);
  useSettingsStore((s) => s.language);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const [showModels, setShowModels] = useState(false);

  useEffect(() => {
    useSettingsStore.getState().refreshModels();
  }, []);

  const reasoningOptions = [
    { key: "low", label: t("reasoning_low"), icon: "\u26a1" },
    { key: "medium", label: t("reasoning_medium"), icon: "\u2696\ufe0f" },
    { key: "high", label: t("reasoning_high"), icon: "\ud83e\udde0" },
  ];
  const reasoningIdx = reasoningOptions.findIndex((o) => o.key === reasoningLevel);
  const modelsMap = (useSettingsStore.getState() as any)._modelsMap as Record<string, { model?: string }> | undefined;

  const close = () => {
    setRawOpen(false);
    setShowModels(false);
  };

  const handleOpen = () => {
    if (rawOpen) {
      close();
      return;
    }
    setRawOpen(true);
  };

  return (
    <div className="relative">
      <button
        ref={triggerRef}
        onClick={handleOpen}
        className={`flex items-center gap-1 px-2 py-1 rounded-lg text-[11px] transition-all cursor-pointer ${
          rawOpen
            ? "bg-accent/15 text-accent border border-accent/30"
            : "text-txt-2 hover:text-txt hover:bg-bdr/40 border border-transparent"
        }`}
        title={t("current_model")}
      >
        <Cpu className="w-3 h-3" />
        <span className="max-w-[120px] truncate font-medium">{currentModel || "..."}</span>
        <ChevronDown className={`w-2.5 h-2.5 transition-transform ${rawOpen ? "rotate-180" : ""}`} />
      </button>

      {open && <div className="fixed inset-0 z-40" onClick={close} />}
      {open && (
        <div className={`absolute bottom-full mb-2 right-0 z-50 ${closing ? "animate-pop-out" : "animate-pop-up"}`}>
          <div className="flex items-end gap-0.5">
            {showModels && (
              <div className="w-[220px] bg-surface border border-bdr rounded-xl shadow-2xl overflow-hidden order-1 animate-slide-right">
                <div className="px-3 pt-2.5 pb-1">
                  <div className="text-[10px] text-txt-2 uppercase tracking-wider mb-1.5">{t("current_model")}</div>
                </div>
                <div className="max-h-[200px] overflow-y-auto px-1.5 pb-1.5 space-y-0.5">
                  {models.map((m) => {
                    const display = modelsMap?.[m]?.model || m;
                    const isActive = m === currentModelKey;
                    return (
                      <button
                        key={m}
                        onClick={() => {
                          useSettingsStore.getState().setCurrentModel(m);
                          close();
                        }}
                        className={`w-full flex items-center justify-between px-2.5 py-2 rounded-lg transition-all cursor-pointer ${
                          isActive ? "bg-accent/10" : "hover:bg-elevated"
                        }`}
                      >
                        <span className={`text-xs font-medium truncate ${isActive ? "text-accent" : "text-txt"}`}>
                          {display}
                        </span>
                        {isActive && <Check className="w-3.5 h-3.5 text-accent flex-shrink-0" />}
                      </button>
                    );
                  })}
                </div>
              </div>
            )}

            <div className={`${showModels ? "order-2" : ""} w-[260px] bg-surface border border-bdr rounded-xl shadow-2xl overflow-hidden`}>
              <div className="px-3.5 pt-3 pb-2.5">
                <div className="flex items-center gap-1.5 mb-2.5">
                  <Brain className="w-3.5 h-3.5 text-accent" />
                  <span className="text-[11px] font-medium text-txt">{t("reasoning_label")}</span>
                </div>
                <div className="relative flex bg-elevated rounded-lg p-0.5">
                  <div
                    className="absolute top-0.5 bottom-0.5 rounded-md bg-accent shadow-sm transition-all duration-200 ease-out"
                    style={{
                      width: `calc(${100 / reasoningOptions.length}% - 2px)`,
                      left: `calc(${(reasoningIdx * 100) / reasoningOptions.length}% + 2px)`,
                    }}
                  />
                  {reasoningOptions.map((opt) => (
                    <button
                      key={opt.key}
                      onClick={() => useSettingsStore.getState().setReasoningLevel(opt.key)}
                      className={`relative z-10 flex-1 flex items-center justify-center gap-1 px-2 py-1.5 text-[11px] rounded-md transition-colors duration-200 cursor-pointer ${
                        opt.key === reasoningLevel ? "text-white font-medium" : "text-txt-2 hover:text-txt"
                      }`}
                    >
                      <span className="text-[10px]">{opt.icon}</span>
                      <span>{opt.label}</span>
                    </button>
                  ))}
                </div>
              </div>
              <div className="h-px bg-bdr-div mx-3" />
              <button
                onClick={() => setShowModels(!showModels)}
                className="w-full flex items-center justify-between px-3.5 py-2.5 text-xs text-txt cursor-pointer hover:bg-elevated transition-colors"
              >
                <span className="flex items-center gap-2">
                  <Cpu className="w-3.5 h-3.5 text-accent" />
                  <span className="font-medium truncate max-w-[160px]">{currentModel || "..."}</span>
                </span>
                <ChevronRight className={`w-3 h-3 text-txt-2 transition-transform duration-200 ${showModels ? "rotate-180" : ""}`} />
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

