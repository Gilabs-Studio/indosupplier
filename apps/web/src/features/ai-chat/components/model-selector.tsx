"use client";

import { useEffect, useMemo } from "react";
import { ChevronDown, Cpu } from "lucide-react";

import { cn } from "@/lib/utils";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";

import { useAIModels } from "../hooks/use-ai-chat";
import { useAIChatStore } from "../stores/use-ai-chat-store";

export function ModelSelector() {
  const { selectedModel, setSelectedModel } = useAIChatStore();
  const { data: modelsResponse } = useAIModels();

  const models = useMemo(() => modelsResponse?.data ?? [], [modelsResponse?.data]);
  const currentModel = models.find((m) => m.id === selectedModel);
  const displayName = currentModel?.display_name ?? "Select Model";

  // Auto-select default model on first load
  useEffect(() => {
    if (models.length === 0) {
      return;
    }

    const selectedStillAvailable = selectedModel
      ? models.some((model) => model.id === selectedModel)
      : false;

    if (!selectedModel || !selectedStillAvailable) {
      const defaultModel = models.find((m) => m.is_default) ?? models[0];
      setSelectedModel(defaultModel.id);
    }
  }, [selectedModel, models, setSelectedModel]);

  if (models.length === 0) return null;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className="h-6 max-w-40 cursor-pointer gap-1 px-2 text-[10px] text-muted-foreground hover:text-foreground"
        >
          <Cpu className="h-3 w-3 shrink-0" />
          <span className="truncate">{displayName}</span>
          <ChevronDown className="h-3 w-3 shrink-0" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        {models.map((model) => (
          <DropdownMenuItem
            key={model.id}
            className={cn(
              "cursor-pointer flex-col items-start gap-0.5",
              selectedModel === model.id && "bg-accent"
            )}
            onClick={() => setSelectedModel(model.id)}
          >
            <span className="text-xs font-medium">{model.display_name}</span>
            <span className="text-[10px] text-muted-foreground">
              {model.description}
            </span>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
