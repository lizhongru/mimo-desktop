import { useCallback, useEffect, useState } from "react";
import { DistillListCandidates } from "../wails/wailsjs/go/desktop/App";

let cachedSkillCommands: Record<string, string> | null = null;

function pickSkillCommand(skill: { name: string; commands?: string[]; pattern?: string }): string {
  return skill.commands?.[0] || skill.pattern || skill.name;
}

export function useSkillCommands() {
  const [commands, setCommands] = useState<Record<string, string>>(cachedSkillCommands || {});

  const load = useCallback(async () => {
    try {
      const candidates = await DistillListCandidates();
      const next: Record<string, string> = {};
      for (const skill of candidates || []) {
        next[skill.name] = pickSkillCommand(skill);
      }
      cachedSkillCommands = next;
      setCommands(next);
    } catch {
      if (cachedSkillCommands) setCommands(cachedSkillCommands);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  return commands;
}

export function getSkillCommandLabel(skillName: string, commands: Record<string, string>): string {
  return commands[skillName] || skillName;
}
