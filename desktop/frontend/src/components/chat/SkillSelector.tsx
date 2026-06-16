import { useCallback, useEffect, useState } from "react";
import { Check, ChevronDown, Sparkles } from "lucide-react";
import type { desktop } from "../../wails/wailsjs/go/models";
import { DistillListCandidates } from "../../wails/wailsjs/go/desktop/App";
import { getSkillCandidateExplanation, t } from "../../lib/i18n";
import { useAnimatedOpen } from "../../lib/useAnimatedOpen";

interface Props {
  selected: string[];
  onChange: (selected: string[]) => void;
}

export function SkillSelector({ selected, onChange }: Props) {
  const [rawOpen, setRawOpen] = useState(false);
  const [skills, setSkills] = useState<desktop.SkillCandidateInfo[]>([]);
  const { shouldRender, closing } = useAnimatedOpen(rawOpen, 120);

  const loadSkills = useCallback(async () => {
    const list = await DistillListCandidates().catch(() => []);
    setSkills((list || []).filter((candidate) => candidate.enabled));
  }, []);

  useEffect(() => {
    void loadSkills();
  }, [loadSkills]);

  useEffect(() => {
    if (rawOpen) void loadSkills();
  }, [rawOpen, loadSkills]);

  const toggleSkill = (name: string) => {
    if (selected.includes(name)) {
      onChange(selected.filter((item) => item !== name));
    } else {
      onChange([...selected, name]);
    }
  };

  const label = selected.length === 0
    ? t("skill_selector_none")
    : `${selected.length}${t("skill_selector_selected_suffix")}`;

  return (
    <div className="relative">
      <button
        onClick={() => setRawOpen(!rawOpen)}
        className={`flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] transition-colors cursor-pointer ${
          selected.length > 0 ? "text-accent bg-accent/10" : "text-txt-2 hover:text-txt hover:bg-bdr/40"
        }`}
        title={t("skill_selector_label")}
      >
        <Sparkles className="w-3 h-3" />
        <span className="max-w-[82px] truncate">{label}</span>
        <ChevronDown className="w-2.5 h-2.5 text-txt-2" />
      </button>

      {shouldRender && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setRawOpen(false)} />
          <div
            className={`absolute bottom-full mb-1.5 left-0 z-50 w-64 max-h-72 overflow-y-auto bg-surface border border-bdr rounded-lg shadow-xl p-2 ${
              closing ? "animate-pop-out" : "animate-pop-up"
            }`}
          >
            <div className="px-2 py-1 text-[10px] uppercase tracking-wide text-txt-g">
              {t("skill_selector_label")}
            </div>
            {skills.length === 0 ? (
              <div className="px-2 py-3 text-xs text-txt-g">{t("skill_selector_empty")}</div>
            ) : (
              <div className="space-y-1">
                {skills.map((skill) => {
                  const checked = selected.includes(skill.name);
                  return (
                    <button
                      key={skill.name}
                      onClick={() => toggleSkill(skill.name)}
                      className={`w-full flex items-start gap-2 rounded-md px-2 py-2 text-left text-xs transition-colors cursor-pointer ${
                        checked ? "bg-accent/10 text-accent" : "text-txt-2 hover:bg-elevated"
                      }`}
                    >
                      <span className={`mt-0.5 flex h-3.5 w-3.5 items-center justify-center rounded border ${checked ? "border-accent bg-accent text-white" : "border-bdr"}`}>
                        {checked && <Check className="h-3 w-3" />}
                      </span>
                      <span className="min-w-0 flex-1">
                        <span className="block truncate font-medium">{skill.name}</span>
                        <span className="mt-0.5 block line-clamp-2 text-[11px] text-txt-g">{skill.explanation || getSkillCandidateExplanation(skill) || skill.description}</span>
                      </span>
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}
