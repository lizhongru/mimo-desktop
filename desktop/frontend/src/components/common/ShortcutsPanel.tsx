import { X } from "lucide-react";
import { t } from "../../lib/i18n";

interface Props {
  open: boolean;
  onClose: () => void;
}

const shortcuts = [
  { key: "Enter", desc: "shortcut_send" },
  { key: "Shift + Enter", desc: "shortcut_new_line" },
  { key: "Ctrl + N", desc: "shortcut_new_chat" },
  { key: "Ctrl + B", desc: "shortcut_toggle_left" },
  { key: "Ctrl + I", desc: "shortcut_toggle_right" },
  { key: "Ctrl + K", desc: "shortcut_compress" },
  { key: "Esc", desc: "shortcut_escape" },
] as const;

export function ShortcutsPanel({ open, onClose }: Props) {
  return (
    <div className={`modal-overlay ${open ? "is-open" : ""}`} onClick={onClose}>
      <div className="modal-dialog w-[400px] mx-4" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between px-5 py-4 border-b border-bdr-div">
          <h3 className="text-sm font-medium text-txt">{t("shortcuts")}</h3>
          <button onClick={onClose} className="p-1 rounded hover:bg-elevated text-txt-g cursor-pointer">
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="px-5 py-3 space-y-1 max-h-[60vh] overflow-y-auto">
          {shortcuts.map((s) => (
            <div key={s.key} className="flex items-center justify-between py-2">
              <span className="text-sm text-txt-2">{t(s.desc)}</span>
              <kbd className="px-2 py-0.5 text-xs bg-elevated border border-bdr rounded text-txt-m font-mono">
                {s.key}
              </kbd>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
