<script lang="ts">
  import Dropzone from '$components/upload/Dropzone.svelte';
  import FileQueue from '$components/upload/FileQueue.svelte';
  import UploadSettings from '$components/upload/UploadSettings.svelte';
  import UploadSuccess from '$components/upload/UploadSuccess.svelte';

  let files: File[] = $state([]);
  let step: 'drop' | 'queue' | 'settings' | 'success' = $state('drop');
  let batchId: string = $state('');
  let slug: string = $state('');
  let ownerToken: string = $state('');

  function handleFilesSelected(selected: File[]) {
    files = selected;
    step = 'queue';
  }

  function handleComplete(slug_: string, ownerToken_: string) {
    slug = slug_;
    ownerToken = ownerToken_;
    step = 'success';
  }

  function handleBack() {
    files = [];
    step = 'drop';
  }
</script>

<svelte:head>
  <title>Upload - xians</title>
</svelte:head>

<div class="mx-auto flex min-h-dvh max-w-2xl flex-col items-center justify-center px-5 py-12">
  <a href="/" class="font-ui mb-12 text-lg font-bold text-ink hover:opacity-80">xians</a>

  {#if step === 'drop'}
    <div class="mb-10 text-center">
      <p class="font-display text-3xl leading-tight text-ink uppercase">Upload files</p>
      <p class="mt-3 text-sm text-body">Drop your files to get started.</p>
    </div>
    <Dropzone onFilesSelected={handleFilesSelected} />
  {:else if step === 'queue'}
    <div class="text-center">
      <p class="font-display text-2xl leading-tight text-ink uppercase">Files</p>
      <p class="mt-1 text-sm text-body">{files.length} file{files.length !== 1 ? 's' : ''} selected</p>
    </div>
    <div class="mt-8 w-full">
      <FileQueue
        {files}
        onBatchInit={(id: string) => (batchId = id)}
        onBack={handleBack}
        onProceed={() => (step = 'settings')}
      />
    </div>
  {:else if step === 'settings'}
    <UploadSettings {batchId} onComplete={handleComplete} onBack={() => (step = 'queue')} />
  {:else if step === 'success'}
    <UploadSuccess {slug} {ownerToken} onNewUpload={handleBack} />
  {/if}
</div>
