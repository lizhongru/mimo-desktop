import { useState, useCallback } from "react";
import {
  Database,
  History,
  Shield,
  Save,
  RotateCcw,
} from "lucide-react";

interface CheckpointConfig {
  autoCheckpoint: boolean;
  tokenThreshold: number;
  maxCheckpoints: number;
}

interface MemoryConfig {
  ccIndex: boolean;
  searchScoreFloor: number;
}

interface PermissionConfig {
  read: string;
  write: string;
  edit: string;
  bash: string;
}

interface Props {
  onSave?: (config: {
    checkpoint: CheckpointConfig;
    memory: MemoryConfig;
    permission: PermissionConfig;
  }) => void;
}

export function AdvancedSettings({ onSave }: Props) {
  const [checkpoint, setCheckpoint] = useState<CheckpointConfig>({
    autoCheckpoint: true,
    tokenThreshold: 75,
    maxCheckpoints: 10,
  });

  const [memory, setMemory] = useState<MemoryConfig>({
    ccIndex: false,
    searchScoreFloor: 15,
  });

  const [permission, setPermission] = useState<PermissionConfig>({
    read: "allow",
    write: "ask",
    edit: "ask",
    bash: "ask",
  });

  const handleSave = useCallback(() => {
    onSave?.({ checkpoint, memory, permission });
  }, [checkpoint, memory, permission, onSave]);

  return (
    <div className="space-y-6">
      {/* Checkpoint Settings */}
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <History className="w-4 h-4 text-accent" />
          <h3 className="text-sm font-medium">检查点设置</h3>
        </div>

        <div className="space-y-3 pl-6">
          <label className="flex items-center gap-3 text-sm">
            <input
              type="checkbox"
              checked={checkpoint.autoCheckpoint}
              onChange={(e) =>
                setCheckpoint({ ...checkpoint, autoCheckpoint: e.target.checked })
              }
              className="rounded border-bdr"
            />
            <span>自动创建检查点</span>
          </label>

          <div className="flex items-center gap-3 text-sm">
            <span>触发阈值:</span>
            <input
              type="range"
              min="50"
              max="90"
              value={checkpoint.tokenThreshold}
              onChange={(e) =>
                setCheckpoint({
                  ...checkpoint,
                  tokenThreshold: parseInt(e.target.value),
                })
              }
              className="flex-1"
            />
            <span className="w-12 text-right">{checkpoint.tokenThreshold}%</span>
          </div>

          <div className="flex items-center gap-3 text-sm">
            <span>最大检查点数:</span>
            <input
              type="number"
              min="1"
              max="50"
              value={checkpoint.maxCheckpoints}
              onChange={(e) =>
                setCheckpoint({
                  ...checkpoint,
                  maxCheckpoints: parseInt(e.target.value) || 10,
                })
              }
              className="w-20 px-2 py-1 bg-elevated border border-bdr rounded text-sm"
            />
          </div>
        </div>
      </div>

      {/* Memory Settings */}
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <Database className="w-4 h-4 text-accent" />
          <h3 className="text-sm font-medium">记忆设置</h3>
        </div>

        <div className="space-y-3 pl-6">
          <label className="flex items-center gap-3 text-sm">
            <input
              type="checkbox"
              checked={memory.ccIndex}
              onChange={(e) =>
                setMemory({ ...memory, ccIndex: e.target.checked })
              }
              className="rounded border-bdr"
            />
            <span>索引 Claude Code 记忆</span>
          </label>

          <div className="flex items-center gap-3 text-sm">
            <span>搜索分数下限:</span>
            <input
              type="range"
              min="5"
              max="50"
              value={memory.searchScoreFloor}
              onChange={(e) =>
                setMemory({
                  ...memory,
                  searchScoreFloor: parseInt(e.target.value),
                })
              }
              className="flex-1"
            />
            <span className="w-12 text-right">{memory.searchScoreFloor}%</span>
          </div>
        </div>
      </div>

      {/* Permission Settings */}
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <Shield className="w-4 h-4 text-accent" />
          <h3 className="text-sm font-medium">权限设置</h3>
        </div>

        <div className="space-y-3 pl-6">
          {Object.entries(permission).map(([key, value]) => (
            <div key={key} className="flex items-center gap-3 text-sm">
              <span className="w-20">{key}:</span>
              <select
                value={value}
                onChange={(e) =>
                  setPermission({ ...permission, [key]: e.target.value })
                }
                className="px-2 py-1 bg-elevated border border-bdr rounded text-sm"
              >
                <option value="allow">Allow</option>
                <option value="ask">Ask</option>
                <option value="deny">Deny</option>
              </select>
            </div>
          ))}
        </div>
      </div>

      {/* Save button */}
      <div className="flex justify-end gap-2 pt-4 border-t border-bdr">
        <button
          onClick={handleSave}
          className="flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-md text-sm hover:bg-accent/90"
        >
          <Save className="w-4 h-4" />
          保存设置
        </button>
      </div>
    </div>
  );
}
