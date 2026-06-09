import { X, Terminal } from "lucide-react";
import { t } from "../../lib/i18n";

interface Props {
  open: boolean;
  onClose: () => void;
}

export function HelpLogPanel({ open, onClose }: Props) {
  return (
    <div className={`modal-overlay ${open ? "is-open" : ""}`} onClick={onClose}>
      <div className="modal-dialog w-[500px] mx-4" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between px-5 py-4 border-b border-bdr-div">
          <h3 className="text-sm font-medium text-txt flex items-center gap-2">
            <Terminal className="w-4 h-4 text-txt-g" />
            {t("help_log_title")}
          </h3>
          <button onClick={onClose} className="p-1 rounded hover:bg-elevated text-txt-g cursor-pointer">
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="px-5 py-4 text-sm text-txt-2">
          No logs collected in this session yet.
        </div>
      </div>
    </div>
  );
}
