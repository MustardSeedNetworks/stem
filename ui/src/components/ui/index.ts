// biome-ignore-all lint/performance/noBarrelFile: Intentional UI primitive barrel — required for Wave 5 Storybook coverage (#236) and cross-repo reuse.

/**
 * UI Component Library
 *
 * Centralized exports for all reusable UI primitives. Consumers should
 * `import { Modal, Button, ... } from '@/components/ui'` rather than
 * deep-importing from the individual files. Required for Wave 5
 * Storybook coverage (#236) and for any future cross-repo reuse.
 */

// Alerts
export type { AlertProps, AlertStatus } from './Alert';
export { Alert } from './Alert';

// Buttons
export { Button, IconButton } from './Button';

// Cards
export type { Status } from './Card';
export { Card, CardDivider, CardRow, CardValue } from './Card';

// Combobox + command palette
export type { ComboboxProps } from './Combobox';
export { Combobox } from './Combobox';
export type { CommandPaletteAction, CommandPaletteProps } from './CommandPalette';
export { CommandPalette } from './CommandPalette';

// Modals
export type { ConfirmModalProps } from './ConfirmModal';
export { ConfirmModal } from './ConfirmModal';
// Inputs + forms
export {
  Checkbox,
  FormGroup,
  FormSection,
  Input,
  SearchInput,
  Select,
  Textarea,
  Toggle,
} from './Input';
export type { InputModalProps } from './InputModal';
export { InputModal } from './InputModal';
export type { ModalProps, ModalSize } from './Modal';
export { Modal, ModalBody, ModalFooter, ModalHeader } from './Modal';
export type { StatusType } from './StatusBadge';
// Status + badges
export { StatusBadge } from './StatusBadge';
export type { SizeKey, Status as StatusKey } from './StatusConfig';
export { getSizeConfig, getStatusConfig, sizeConfig, statusConfig } from './StatusConfig';

// Tags
export { Tag } from './Tag';

// Tooltips
export type { TooltipProps } from './Tooltip';
export { Tooltip } from './Tooltip';

// Typography
export {
  AccentLink,
  Caption,
  H1,
  H2,
  H3,
  H4,
  P,
  SmallText,
} from './Typography';
