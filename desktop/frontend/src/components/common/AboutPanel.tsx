import { useEffect, useState } from "react";
import { X, Github } from "lucide-react";
import { t } from "../../lib/i18n";

interface Props {
  open: boolean;
  onClose: () => void;
}

export function AboutPanel({ open, onClose }: Props) {
  const [version, setVersion] = useState<{ appVersion?: string }>({});

  useEffect(() => {
    if (!open) return;
    window.go?.desktop?.App?.GetVersion?.()
      .then((v) => setVersion(v))
      .catch(console.error);
  }, [open]);

  return (
    <div className={`modal-overlay ${open ? "is-open" : ""}`} onClick={onClose}>
      <div className="modal-dialog w-[380px] mx-4" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between px-5 py-4 border-b border-bdr-div">
          <h3 className="text-sm font-medium text-txt">{t("about_title")}</h3>
          <button onClick={onClose} className="p-1 rounded hover:bg-elevated text-txt-g cursor-pointer">
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="px-5 py-4 space-y-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-accent/20 flex items-center justify-center">
              <span className="text-lg font-bold text-accent">M</span>
            </div>
            <div>
              <div className="text-sm font-medium text-txt">MiMo Desktop</div>
              <div className="text-xs text-txt-g">{version.appVersion || "v0.4.0-dev"}</div>
            </div>
          </div>
          <p className="text-xs text-txt-2">{t("about_description")}</p>
          <a
            href="https://github.com/lizhongru/mimo-desktop"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 px-3 py-2 rounded-lg bg-elevated hover:bg-elevated/80 transition-colors text-xs text-txt-2 cursor-pointer"
          >
            <Github className="w-3.5 h-3.5" />
            {t("about_github")}
          </a>
        </div>
      </div>
    </div>
  );
}
