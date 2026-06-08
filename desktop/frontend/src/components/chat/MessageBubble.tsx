import { User, Bot } from "lucide-react";
import type { ChatMessage } from "../../lib/types";
import { MarkdownRenderer } from "./MarkdownRenderer";
import { ThinkingBlock } from "./ThinkingBlock";
import { ToolCallCard } from "./ToolCallCard";

interface Props {
  message: ChatMessage;
}

export function MessageBubble({ message }: Props) {
  const isUser = message.role === "user";

  return (
    <div className={`flex gap-3 py-3 ${isUser ? "justify-end" : ""}`}>
      {!isUser && (
        <div className="w-7 h-7 rounded-full bg-accent/20 flex items-center justify-center flex-shrink-0 mt-0.5">
          <Bot className="w-4 h-4 text-accent" />
        </div>
      )}

      <div
        className={`max-w-[80%] rounded-lg px-4 py-2.5 text-sm ${
          isUser
            ? "bg-accent text-white"
            : "bg-elevated/60 text-txt"
        }`}
      >
        {isUser ? (
          <p className="whitespace-pre-wrap">{message.content}</p>
        ) : (
          <div>
            {/* Completed thinking */}
            {message.thinking && (
              <ThinkingBlock content={message.thinking} />
            )}

            {/* Completed tool calls */}
            {message.toolCalls?.map((tc) => (
              <ToolCallCard key={tc.id} toolCall={tc} />
            ))}

            {/* Completed content */}
            <MarkdownRenderer content={message.content} />
          </div>
        )}

        {/* Duration for assistant messages */}
        {!isUser && message.duration && (
          <div className="mt-1.5 text-[10px] text-txt-m">
            {(message.duration / 1000).toFixed(1)}s
          </div>
        )}
      </div>

      {isUser && (
        <div className="w-7 h-7 rounded-full bg-elevated flex items-center justify-center flex-shrink-0 mt-0.5">
          <User className="w-4 h-4 text-txt-2" />
        </div>
      )}
    </div>
  );
}
