import { X, History } from "lucide-react";
import { t } from "../../lib/i18n";
import { CheckpointPanel } from "../checkpoint/CheckpointPanel";

interface Props {
  open: boolean;
  onClose: () => void;
}

export function CheckpointPanelModal({ open, onClose }: Props) {
  return (
    <div className={`modal-overlay ${open ? "is-open" : ""}`} onClick={onClose}>
      <div className="modal-dialog w-[500px] mx-4" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between px-5 py-4 border-b border-bdr-div">
          <h3 className="text-sm font-medium text-txt flex items-center gap-2">
            <History className="w-4 h-4 text-txt-g" />
            {t("checkpoint") || "检查点管理"}
          </h3>
          <button onClick={onClose} className="p-1 rounded hover:bg-elevated text-txt-g cursor-pointer">
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="overflow-y-auto max-h-[60vh]">
          <CheckpointPanel />
        </div>
      </div>
    </div>
  );
}
