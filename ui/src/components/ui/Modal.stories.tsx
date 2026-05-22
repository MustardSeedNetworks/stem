/**
 * Modal primitive stories (Wave 5 / #236).
 *
 * Covers the size matrix, the showCloseButton/closeOnBackdropClick/
 * closeOnEscape flag combinations, and the with-title vs no-title
 * shells.
 */
import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { Button, Modal, ModalBody, ModalFooter, ModalHeader } from './';

const meta: Meta<typeof Modal> = {
  title: 'UI/Modal',
  component: Modal,
  parameters: { layout: 'fullscreen' },
  argTypes: {
    size: { control: 'select', options: ['sm', 'md', 'lg', 'xl', 'full'] },
    isOpen: { control: 'boolean' },
    showCloseButton: { control: 'boolean' },
    closeOnBackdropClick: { control: 'boolean' },
    closeOnEscape: { control: 'boolean' },
  },
};
export default meta;

type Story = StoryObj<typeof Modal>;

const SampleBody = () => (
  <ModalBody>
    <p>Modal content goes here. This body uses ModalBody for consistent spacing.</p>
    <p className="text-text-muted text-sm">A second paragraph to show vertical rhythm.</p>
  </ModalBody>
);

const SampleFooter = ({ onClose }: { onClose: () => void }) => (
  <ModalFooter>
    <Button variant="ghost" onClick={onClose}>
      Cancel
    </Button>
    <Button variant="solid" tone="violet" onClick={onClose}>
      Confirm
    </Button>
  </ModalFooter>
);

export const Default: Story = {
  args: {
    isOpen: true,
    title: 'Default modal',
    onClose: () => undefined,
    children: <SampleBody />,
  },
};

export const WithFooter: Story = {
  args: { isOpen: true, title: 'Modal with footer', onClose: () => undefined },
  render: function withFooterRender(args) {
    const [open, setOpen] = useState(args.isOpen);
    const close = () => setOpen(false);
    return (
      <Modal {...args} isOpen={open} onClose={close}>
        <SampleBody />
        <SampleFooter onClose={close} />
      </Modal>
    );
  },
};

export const NoTitleNoCloseButton: Story = {
  args: {
    isOpen: true,
    showCloseButton: false,
    onClose: () => undefined,
    children: <SampleBody />,
  },
};

export const SmallSize: Story = {
  args: {
    isOpen: true,
    size: 'sm',
    title: 'Small',
    onClose: () => undefined,
    children: <SampleBody />,
  },
};
export const LargeSize: Story = {
  args: {
    isOpen: true,
    size: 'lg',
    title: 'Large',
    onClose: () => undefined,
    children: <SampleBody />,
  },
};
export const FullSize: Story = {
  args: {
    isOpen: true,
    size: 'full',
    title: 'Full width',
    onClose: () => undefined,
    children: <SampleBody />,
  },
};

export const InteractiveToggle: Story = {
  args: {
    isOpen: false,
    title: 'Trigger demo',
    onClose: () => undefined,
    children: <SampleBody />,
  },
  render: function interactiveRender(args) {
    const [open, setOpen] = useState(false);
    return (
      <div className="p-8">
        <Button onClick={() => setOpen(true)}>Open modal</Button>
        <Modal {...args} isOpen={open} onClose={() => setOpen(false)}>
          <ModalHeader>
            <h2 className="text-lg font-semibold">{args.title}</h2>
          </ModalHeader>
          <SampleBody />
          <SampleFooter onClose={() => setOpen(false)} />
        </Modal>
      </div>
    );
  },
};
