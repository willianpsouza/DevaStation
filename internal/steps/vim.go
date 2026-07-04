package steps

import (
	"devstation/internal/step"
	"devstation/internal/system"
)

// Vim installs vim and drops a sensible, dependency-free vimrc tuned for Go.
type Vim struct{}

func (Vim) ID() string    { return "vim" }
func (Vim) Title() string { return "vim + vimrc p/ Go" }

func (Vim) Check(c *step.Context) (bool, error) {
	if !system.HasCommand("vim") {
		return false, nil
	}
	return fileHasMarker(c.Target.Home+"/.vimrc", managedMarker), nil
}

func (Vim) Apply(c *step.Context) error {
	if err := c.AptInstall("vim"); err != nil {
		return err
	}
	vimrc := managedMarker + `
" ~/.vimrc — perfil dev Go (sem gerenciador de plugins, só stdlib do vim)

set nocompatible
filetype plugin indent on
syntax on

" --- Interface ---
set number relativenumber
set ruler
set showcmd
set cursorline
set laststatus=2
set wildmenu
set wildmode=longest:full,full
set scrolloff=4
set signcolumn=yes
set termguicolors
set background=dark

" --- Edição ---
set expandtab
set shiftwidth=4
set tabstop=4
set softtabstop=4
set autoindent
set smartindent
set backspace=indent,eol,start
set clipboard=unnamedplus

" --- Busca ---
set ignorecase
set smartcase
set incsearch
set hlsearch

" --- Comportamento ---
set hidden
set mouse=a
set updatetime=300
set undofile
set undodir=~/.vim/undo
set directory=~/.vim/swap//
set backupdir=~/.vim/backup//

" Cria os diretórios de estado no primeiro uso.
silent! call mkdir(expand('~/.vim/undo'), 'p')
silent! call mkdir(expand('~/.vim/swap'), 'p')
silent! call mkdir(expand('~/.vim/backup'), 'p')

" --- Go: usa tabs reais (padrão gofmt) ---
augroup golang
  autocmd!
  autocmd FileType go setlocal noexpandtab tabstop=4 shiftwidth=4
  " gofmt ao salvar (usa o binário do sistema, sem plugin)
  autocmd BufWritePre *.go silent! %!gofmt
augroup END

" YAML/JSON com 2 espaços
autocmd FileType yaml,json setlocal expandtab shiftwidth=2 tabstop=2 softtabstop=2

" --- Atalhos ---
let mapleader = " "
nnoremap <leader>w :w<CR>
nnoremap <leader>q :q<CR>
nnoremap <leader>h :nohlsearch<CR>
nnoremap <C-n> :bnext<CR>
nnoremap <C-p> :bprev<CR>
`
	if err := c.WriteUserFile(c.Target, c.Target.Home+"/.vimrc", vimrc, 0o644); err != nil {
		return err
	}
	c.UI.Info("~/.vimrc gerado")
	return nil
}
