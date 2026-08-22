export type ActionResult = {
  error?: string;
  ok?: boolean;
};

/**
 * Converte uma server action que retorna resultado em uma action compatível
 * com o atributo `action` de formulários (que exige void ou Promise<void>).
 */
export function voidAction(
  action: (formData: FormData) => Promise<ActionResult>
): (formData: FormData) => Promise<void> {
  return async (formData: FormData) => {
    await action(formData);
  };
}