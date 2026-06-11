export interface ChatAttachment {
  name: string;
  type: string;
  dataUrl: string;
}

export function attachmentNameForFile(file: File): string {
  const relativePath = (file as File & { webkitRelativePath?: string }).webkitRelativePath;
  return relativePath || file.name;
}

export function readFileAsAttachment(file: File): Promise<ChatAttachment> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      resolve({
        name: attachmentNameForFile(file),
        type: file.type || "application/octet-stream",
        dataUrl: reader.result as string,
      });
    };
    reader.onerror = () => reject(reader.error || new Error("Failed to read file"));
    reader.readAsDataURL(file);
  });
}

export async function readFilesAsAttachments(files: FileList | File[]): Promise<ChatAttachment[]> {
  return Promise.all(Array.from(files).map(readFileAsAttachment));
}

