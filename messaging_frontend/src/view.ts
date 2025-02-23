import { WorkspaceItem, WorkspacePage } from "./workspace";
import { ChannelItem } from "./channel";
import { PostItem } from "./posts";
import {
  CreateEvent,
  LoginEvent,
  openEvent,
  openEventChannel,
  refreshChannels,
  refreshWorkspaces,
  LogoutEvent,
  ViewPost,
  viewPostHierarhcy,
  ViewChannel,
  DeleteChannel,
  ViewWorkspace,
  DeleteWorkspace,
  createChannelEvent,
  createWorkspaceEvent,
  messageEvent,
  replytoPost,
  sendButton,
  workspaceDeleted,
  channelDeleted,
  subscribeEvent,
  abortEvent,
  patchEvent,
} from "./types";
import { LoginPage } from "./login";
customElements.define("workspace-page", WorkspacePage);
customElements.define("workspace-item", WorkspaceItem);
customElements.define("channel-item", ChannelItem);
customElements.define("post-item", PostItem);
customElements.define("login-page", LoginPage);

declare global {
  interface DocumentEventMap {
    createEvent: CustomEvent<CreateEvent>;
    deleteWorkspace: CustomEvent<DeleteWorkspace>;
    workspaceDeleted: CustomEvent<workspaceDeleted>;
    channelDeleted: CustomEvent<channelDeleted>;
    deleteChannel: CustomEvent<DeleteChannel>;
    loginEvent: CustomEvent<LoginEvent>;
    openEvent: CustomEvent<openEvent>;
    openEventChannel: CustomEvent<openEventChannel>;
    refreshWorkspaces: CustomEvent<refreshWorkspaces>;
    refreshChannels: CustomEvent<refreshChannels>;
    abortEvent: CustomEvent<abortEvent>;
    logoutEvent: CustomEvent<LogoutEvent>;
    createChannelEvent: CustomEvent<createChannelEvent>;
    createWorkspaceEvent: CustomEvent<createWorkspaceEvent>;
    messageEvent: CustomEvent<messageEvent>;
    replytoPost: CustomEvent<replytoPost>;
    sendButton: CustomEvent<sendButton>;
    subscribeEvent: CustomEvent<subscribeEvent>;
    patchEvent: CustomEvent<patchEvent>;
  }
}

export class View {
  private _openWorkspaceId: string | null = null;
  private _openChannelId: string | null = null;
  private _openWorkspaces: Set<string> = new Set();
  private _openChannels: Set<string> = new Set();
  private _openPosts: Set<string> = new Set();
  private _postsContainer: any;
  private _channelsContainer: any;
  private _workspaceContainer: any;
  private _parent: string | undefined = undefined;
  private _username: string | null = null;
  private _postlists: any;
  private _page: WorkspacePage | undefined;
  private _login: LoginPage | undefined;
  private _sending: boolean = false;

  /**
   * Returns the username of the user logged on
   * @returns the username of the user
   */
  getUsername(): string {
    return this._username ?? "temp";
  }

  /**
   * Change the parent of the text box for a reply
   * @param parent The parent now being replied to
   */
  setParent(parent: string | null = null): void {
    if (this._parent) {
      // Delete previous text box at a reply
      const postElement = this._postsContainer?.querySelector(
        `[data-post-id="${this._parent}"]`,
      ) as PostItem;
      postElement?.removeTextBox();
    } else {
      // console.log("Deleting default text box");
      // Delete previous text box at the bottom
      const oldMessage = this._postsContainer?.querySelector("textarea");
      if (oldMessage) {
        oldMessage.remove();
      }
      let buttonsLeft = true;
      while (buttonsLeft) {
        const button = this._postsContainer?.querySelector("button");
        if (button) {
          button.remove();
        } else {
          buttonsLeft = false;
        }
      }
    }

    if (parent) {
      // Moving textbox to parent
      this._parent = parent;
      console.log("parent is now ", this._parent);

      const textBox = this._postsContainer?.querySelector(
        `[data-post-id="${parent}"]`,
      ) as PostItem;
      textBox.addTextBox(this._openChannelId, this._parent);
      textBox.scrollIntoView({ behavior: "smooth", block: "start" });
    } else {
      //Moving textbox to default position
      const textBox = this._postsContainer.querySelector("form");

      const message = document.createElement("textarea");
      message.id = "new-message";
      message.name = "message";
      message.placeholder = "Type your message here...";
      textBox.append(message);
      let shiftPressed = false;

      //Go to new line if shift is pressed and submit post otherwise
      message?.addEventListener("keydown", (event) => {
        // console.log("key is " + event.key);
        if (event.key == "Shift") {
          shiftPressed = true;
        }

        if (event.key == "Enter") {
          // console.log("Enter pressed");
          if (!shiftPressed) {
            // console.log("Showing post input");
            const messageEvent = new CustomEvent("messageEvent", {
              detail: {
                name: this._openChannelId,
                message: message.value,
                parent: this._parent,
              },
            });
            document.dispatchEvent(messageEvent);
          } else {
            // console.log("Enter shift");
          }
        }
      });

      message?.addEventListener("keyup", (event) => {
        if (event.key == "Shift") {
          shiftPressed = false;
        }
      });

      //Create button for send that sends posts
      const sendButton = document.createElement("button");
      sendButton.type = "button";
      sendButton.id = "send-button";
      sendButton.textContent = "Send";
      textBox.append(sendButton);

      sendButton?.addEventListener("click", () => {
        // console.log("Send clicked");
        // console.log("Showing post input");
        if (!this._sending) {
          const messageEvent = new CustomEvent("messageEvent", {
            detail: {
              name: this._openChannelId,
              message: message.value,
              parent: this._parent,
            },
          });
          this._sending = true;
          document.dispatchEvent(messageEvent);
        } else {
          console.log("Message is already being sent!");
        }
      });

      //Create button for smile that adds a smile format
      const smileButton = document.createElement("button");
      smileButton.type = "button";
      smileButton.id = "smile-button";
      const smileIcon = document.createElement("iconify-icon");
      smileIcon.setAttribute("icon", "emojione:smiling-face");
      smileIcon.setAttribute("width", "1.25em");
      smileIcon.setAttribute("height", "1.25em");
      smileButton.append(smileIcon);
      textBox.append(smileButton);

      smileButton?.addEventListener("click", () => {
        // console.log("Smile clicked");
        message.value = message.value.concat(":smile:");
      });

      //Create button for frown that adds a frown format
      const frownButton = document.createElement("button");
      frownButton.type = "button";
      frownButton.id = "frown-button";
      const frownIcon = document.createElement("iconify-icon");
      frownIcon.setAttribute("icon", "emojione:frowning-face");
      frownIcon.setAttribute("width", "1.25em");
      frownIcon.setAttribute("height", "1.25em");
      frownButton.append(frownIcon);
      textBox.append(frownButton);

      frownButton?.addEventListener("click", () => {
        // console.log("Frown clicked");
        message.value = message.value.concat(":frown:");
      });

      //Create button for like that adds a like format
      const likeButton = document.createElement("button");
      likeButton.type = "button";
      likeButton.id = "like-button";
      const likeIcon = document.createElement("iconify-icon");
      likeIcon.setAttribute("icon", "emojione:thumbs-up");
      likeIcon.setAttribute("width", "1.25em");
      likeIcon.setAttribute("height", "1.25em");
      likeButton.append(likeIcon);
      textBox.append(likeButton);

      likeButton?.addEventListener("click", () => {
        // console.log("Like clicked");
        message.value = message.value.concat(":like:");
      });

      //Create button for celebrate that adds a celebrate format
      const celebrateButton = document.createElement("button");
      celebrateButton.type = "button";
      celebrateButton.id = "celebrate-button";
      const celebrateIcon = document.createElement("iconify-icon");
      celebrateIcon.setAttribute("icon", "emojione:party-popper");
      celebrateIcon.setAttribute("width", "1.25em");
      celebrateIcon.setAttribute("height", "1.25em");
      celebrateButton.append(celebrateIcon);
      textBox.append(celebrateButton);

      celebrateButton?.addEventListener("click", () => {
        // console.log("Celebrate clicked");
        message.value = message.value.concat(":celebrate:");
      });

      //Create button for bold that adds a bold format
      const boldButton = document.createElement("button");
      boldButton.type = "button";
      boldButton.id = "bold-button";
      boldButton.textContent = "bold";
      textBox.append(boldButton);

      boldButton?.addEventListener("click", () => {
        // console.log("Bold clicked");
        // console.log("Selected text index start:", message.selectionStart);
        // console.log("Selected text index end:", message.selectionEnd);
        message.value = message.value
          .substring(0, message.selectionStart)
          .concat(
            "**",
            message.value.substring(
              message.selectionStart,
              message.selectionEnd,
            ),
            "**",
            message.value.substring(message.selectionEnd),
          );
      });

      //Create button for ital that adds a ital format
      const italicButton = document.createElement("button");
      italicButton.type = "button";
      italicButton.id = "italic-button";
      italicButton.textContent = "italic";
      textBox.append(italicButton);

      italicButton?.addEventListener("click", () => {
        // console.log("Italic clicked");
        message.value = message.value
          .substring(0, message.selectionStart)
          .concat(
            "*",
            message.value.substring(
              message.selectionStart,
              message.selectionEnd,
            ),
            "*",
            message.value.substring(message.selectionEnd),
          );
      });

      //Create a button for link that adds a link format
      const linkButton = document.createElement("button");
      linkButton.type = "button";
      linkButton.id = "link-button";
      linkButton.textContent = "link";
      textBox.append(linkButton);

      linkButton?.addEventListener("click", () => {
        // console.log("Link clicked");
        message.value = message.value
          .substring(0, message.selectionStart)
          .concat(
            "[",
            message.value.substring(
              message.selectionStart,
              message.selectionEnd,
            ),
            "]",
            "(",
            message.value.substring(
              message.selectionStart,
              message.selectionEnd,
            ),
            ")",
            message.value.substring(message.selectionEnd),
          );
      });

      // Create a button for superscript that runs a superscript format
      const superscriptButton = document.createElement("button");
      superscriptButton.type = "button";
      superscriptButton.id = "superscript-button";
      superscriptButton.textContent = "superscript";
      textBox.append(superscriptButton);

      superscriptButton?.addEventListener("click", () => {
        // console.log("Superscript clicked");
        message.value = message.value
          .substring(0, message.selectionStart)
          .concat(
            "^^",
            message.value.substring(
              message.selectionStart,
              message.selectionEnd,
            ),
            "^^",
            message.value.substring(message.selectionEnd),
          );
      });

      //Create a button for subscript that runs a subscript format
      const subscriptButton = document.createElement("button");
      subscriptButton.type = "button";
      subscriptButton.id = "subscript-button";
      subscriptButton.textContent = "subscript";
      textBox.append(subscriptButton);

      subscriptButton?.addEventListener("click", () => {
        // console.log("Subscript clicked");
        message.value = message.value
          .substring(0, message.selectionStart)
          .concat(
            "__",
            message.value.substring(
              message.selectionStart,
              message.selectionEnd,
            ),
            "__",
            message.value.substring(message.selectionEnd),
          );
      });
    }
  }

  constructor() {}

  /**
   * Display loggin prompt
   */
  displayPage(username: string): void {
    const mainContainer = document.getElementById("main-container");
    // console.log("calling new template");
    const WorkspacePage = document.createElement(
      "workspace-page",
    ) as WorkspacePage;
    WorkspacePage.setAttribute("workspacePage-id", "workspacePage"); // Use a unique identifier for each channel
    mainContainer?.append(WorkspacePage);
    this.displayLoggedInUser(username);
    this._channelsContainer = WorkspacePage.channelsSection;
    this._postsContainer = WorkspacePage.postsSection;
    this._workspaceContainer = WorkspacePage.workspaceSection;
    this._postlists = WorkspacePage.postlist;
    this._page = WorkspacePage;
  }

  /**
   * Display a login error
   */
  displayLoginError(error: string): void {
    this._login?.displayError();
  }

  /**
   * Display a workspace error
   */
  dispalyWorkspaceError(error: string): void {
    this._page?.displayWorkspaceError(error);
  }

  /**
   * Display a channel error
   */
  dispalyChannelError(error: string): void {
    this._page?.displayChannelError(error);
  }

  /**
   * Display loggin promt
   */
  displayLoginPrompt(): void {
    const header = document.querySelector("#logout");
    // Assuming you have a reference to the specific WorkspacePage element you want to remove
    const workspacePageToRemove = document.querySelector(
      "workspace-page[workspacePage-id='workspacePage']",
    );

    // Check if the element exists before attempting to remove it
    if (workspacePageToRemove) {
      workspacePageToRemove.remove();
    }
    // console.log("calling login template");
    this._login = document.createElement("login-page") as LoginPage;
    header?.append(this._login);
  }

  /**
   * Display logged in user
   *
   * @param username name of currently logged in user
   */
  displayLoggedInUser(username: string): void {
    const logoutHandler = (event: MouseEvent): void => {
      // console.log("Clicked logout button");
      const logoutEvent = new CustomEvent("logoutEvent", {
        detail: { message: "enteredText " },
      });
      const logoutForm = document.getElementById("logoutForm");
      if (logoutForm) {
        logoutForm.remove();
        logoutForm.removeEventListener("click", logoutHandler);
      }
      document.dispatchEvent(logoutEvent);
      event.preventDefault();
    };

    const loginPageToRemove = document.querySelector("login-page");
    if (loginPageToRemove) {
      loginPageToRemove.remove();
    }

    const header = document.querySelector("#logout");

    const nameholder = document.createElement("strong");
    nameholder.textContent = " Currently logged in  (" + username + ")"; // Added a space before text for separation
    this._username = username;
    nameholder.style.marginRight = "10px"; // Add some margin to separate text and button

    const newForm = document.createElement("form");
    newForm.id = "logoutForm";
    const logout = document.createElement("button");
    logout.textContent = "Logout";
    newForm.append(logout, nameholder);
    header?.append(newForm);

    // Handler for a loginEvent
    logout.addEventListener("click", logoutHandler);
  }

  /**
   * Display all Workspaces
   *
   * @param workspaces array of workspaces (documents in the database) to display
   */
  displayAllWorkspaces(
    workspaces: Array<ViewWorkspace>,
    refreshbutton: boolean,
  ): void {
    let navbar = this._workspaceContainer.querySelector("nav");

    if (refreshbutton) {
      this.closeAllWorkspaces();
    } else {
      // create refresh button
      const refreshIcon = document.createElement("iconify-icon");
      refreshIcon.setAttribute("icon", "ic:outline-refresh");
      refreshIcon.setAttribute("width", "1.25em");
      refreshIcon.setAttribute("height", "1.25em");
      const refresh = document.createElement("button");
      refresh.setAttribute("aria-label", "refresh all workspaces");
      refresh.append(refreshIcon);
      navbar.append(refresh);

      //   Rrefresh Handler
      refresh.addEventListener("click", () => {
        const refreshWorkspaces = new CustomEvent("refreshWorkspaces", {
          detail: { name: "path" },
        });
        // console.log("Clicked on rereshing workspaces");
        document.dispatchEvent(refreshWorkspaces);
      });
    }

    this._page?.removeWorkspaceErrors();

    if (workspaces.length === 0) {
      const error = document.createElement("p");
      error.textContent = "no workspaces yet..";
      error.setAttribute("id", "currError");
      navbar.append(error);
    } else {
      // add all the workspaces
      workspaces.forEach((currWorkspace: ViewWorkspace) => {
        this._openWorkspaces.add(currWorkspace.path);

        const workspaceItem = document.createElement(
          "workspace-item",
        ) as WorkspaceItem;
        workspaceItem.data = currWorkspace;
        workspaceItem.setAttribute("data-workspace-id", currWorkspace.path);
        navbar.append(workspaceItem);
        // After creating the workspace item, update its active state
      });
    }
  }

  /**
   * Display all channels in Workspace
   *
   * @param channels array of workspaces (documents in the database) to display
   */
  displayChannels(
    channels: Array<ViewChannel>,
    workspace: string,
    refreshing: boolean,
  ): void {
    this._page?.removeWorkspaceErrors();
    this._page?.removeChannelErrors();
    if (refreshing) {
      // if refreshing the channels check if its already open
      if (this._openWorkspaceId === workspace) {
        // console.log("refreshing the current workspace");
      } else {
        // console.log("this workspace is not open, disregard request");
        return;
      }
    }

    this.closeAllChannels();

    // create refresh button
    const refreshIcon = document.createElement("iconify-icon");
    refreshIcon.setAttribute("icon", "ic:outline-refresh");
    refreshIcon.setAttribute("width", "1.25em");
    refreshIcon.setAttribute("height", "1.25em");
    const refresh = document.createElement("button");
    refresh.setAttribute("id", "channelsrefresh");
    refresh.setAttribute("aria-label", "Refresh Channels");
    refresh.append(refreshIcon);
    this._channelsContainer.append(refresh);

    //   Rrefresh Handler
    refresh.addEventListener("click", () => {
      const refreshChannels = new CustomEvent("refreshChannels", {
        detail: { name: this._openWorkspaceId },
      });
      // console.log("Clicked on rereshing workspaces");
      document.dispatchEvent(refreshChannels);
    });

    this._openWorkspaceId = workspace;

    const prevWorkspace = this._workspaceContainer.querySelector(
      `[data-workspace-id="${this._openWorkspaceId}"]`,
    );
    if (prevWorkspace) {
      prevWorkspace.ActiveState();
    }

    // Create the form element
    const formElement = document.createElement("form");
    formElement.id = "CreateChannelForm";

    // Create the input element
    const input = document.createElement("input");
    input.type = "text";
    input.id = "username";
    input.placeholder = "Create channel";

    // Append the input to the form
    formElement.appendChild(input);

    formElement?.addEventListener("submit", (event: Event) => {
      event.preventDefault();

      const createChannelEvent = new CustomEvent("createChannelEvent", {
        detail: { workspace: this._openWorkspaceId, channel: input?.value },
      });
      // Notification of login event
      document.dispatchEvent(createChannelEvent);
      // formElement.remove();
      input.value = "";
    });

    this._channelsContainer.appendChild(formElement);

    // if any error -->  remove it
    const currerror =
      this._channelsContainer?.querySelector(`[id = "currError"]`);

    if (currerror) {
      currerror.remove();
    }

    const list = document.createElement("ul");
    this._channelsContainer.appendChild(list);

    if (channels.length === 0) {
      const error = document.createElement("li");
      error.style.listStyle = "none";
      error.textContent = "no channels yet..";
      error.setAttribute("id", "currError");
      list.append(error);
    } else {
      channels.forEach((currChannel: ViewChannel) => {
        const wrapper = document.createElement("li");
        wrapper.style.listStyle = "none";
        // Add channel ID to the open channels set so we can close them later
        this._openChannels.add(currChannel.path);

        const channelItem = document.createElement(
          "channel-item",
        ) as ChannelItem;
        channelItem.data = currChannel;
        channelItem.setAttribute("data-channel-id", currChannel.path); // Use a unique identifier for each channel
        wrapper.appendChild(channelItem);
        list.append(wrapper);
      });
    }
  }

  // Method to close all channels
  closeAllWorkspaces(): void {
    this.closeAllChannels();
    // if any error -->  remove it
    const currerror =
      this._workspaceContainer?.querySelector(`[id = "currError"]`);

    if (currerror) {
      currerror.remove();
    }

    this._openWorkspaces.forEach((workspaceId) => {
      // go through the workspaces and remove each
      const workspaceElement = this._workspaceContainer?.querySelector(
        `[data-workspace-id="${workspaceId}"]`,
      );
      if (workspaceElement) {
        workspaceElement.remove();
      }
    });
  }

  // Method to close all channels
  closeAllChannels(): void {
    const currefresh = this._channelsContainer?.querySelector(
      `[id = "channelsrefresh"]`,
    );

    if (currefresh) {
      currefresh.remove();
    }

    const prevWorkspace = this._workspaceContainer.querySelector(
      `[data-workspace-id="${this._openWorkspaceId}"]`,
    );
    if (prevWorkspace) {
      prevWorkspace.DeactivateState();
    }
    // if any error -->  remove it
    const currbox = this._channelsContainer?.querySelector(
      `[id = "CreateChannelForm"]`,
    );

    // if any error -->  remove it
    const prevlist = this._channelsContainer?.querySelector("ul");

    if (prevlist) {
      prevlist.remove();
    }

    if (currbox) {
      currbox.remove();
    }
    this.closeAllPosts();
    this._openChannels.forEach((channelId) => {
      // go through the channels and remove each
      const channelElement = this._channelsContainer?.querySelector(
        `[data-channel-id="${channelId}"]`,
      );
      if (channelElement) {
        channelElement.remove();
      }
    });
    this._openChannels.clear();
  }

  makePostHierarhcy(posts: Array<ViewPost>): Array<viewPostHierarhcy> {
    var orderedPosts = posts.slice();
    orderedPosts.sort((a, b) => a.meta.createdAt - b.meta.createdAt);
    var hierarhcyPosts: Array<viewPostHierarhcy> = new Array();
    orderedPosts.forEach((currPost: ViewPost) => {
      if (currPost.doc.parent !== undefined) {
        // console.log("has parent!");
        // Find index of parent and insert it there with slashes
        var index = 0;
        var indent = 0;
        var found = false;
        var inside = false;
        hierarhcyPosts.every((hierCurrPost) => {
          if (found == true) {
            if (hierCurrPost.indents < indent) {
              inside = true;
              return false;
            }
          }
          if (hierCurrPost.path == currPost.doc.parent) {
            found = true;
            indent = hierCurrPost.indents + 1;
          }
          index = index + 1;
          return true;
        });
        const hierarhcyPost: viewPostHierarhcy = {
          path: currPost.path,
          meta: currPost.meta,
          doc: currPost.doc,
          indents: indent,
        };
        if (inside) {
          hierarhcyPosts.splice(index, 0, hierarhcyPost);
        } else {
          hierarhcyPosts.push(hierarhcyPost);
        }
      } else {
        // console.log("no parent...");
        const hierarhcyPost: viewPostHierarhcy = {
          path: currPost.path,
          meta: currPost.meta,
          doc: currPost.doc,
          indents: 0,
        };
        hierarhcyPosts.push(hierarhcyPost);
      }
    });
    return hierarhcyPosts;
  }

  showPosts(hierarhcyPosts: Array<viewPostHierarhcy>, channel: string): void {
    this._sending = false;
    this._page?.removeChannelErrors();
    // Update the currently open workspace ID
    this._openChannelId = channel;
    // console.log("updated channel:", this._openChannelId);

    // if any error -->  remove it
    const currerror = this._postsContainer?.querySelector(`[id = "currError"]`);
    let wrapper = document.createElement("li");
    if (currerror) {
      currerror.remove();
    }

    if (hierarhcyPosts.length === 0) {
      const error = document.createElement("li");
      error.textContent = "no posts yet..";
      error.style.listStyle = "none";
      error.setAttribute("id", "currError");
      this._postlists.append(error);
    } else {
      hierarhcyPosts.forEach((currPost: viewPostHierarhcy) => {
        this._openPosts.add(currPost.path);

        const postItem = document.createElement("post-item") as PostItem;
        postItem.data = currPost;
        postItem.setUsername(this._username);
        wrapper = document.createElement("li");
        wrapper.style.listStyle = "none";
        wrapper.appendChild(postItem);
        this._postlists.append(wrapper);
        postItem.setAttribute("data-post-id", currPost.path);
        console.log("post id: ", currPost.path);
      });
    }
    setTimeout(() => {
      wrapper.scrollIntoView({ behavior: "smooth", block: "start" });
    }, 50); // Delay of 100 milliseconds

    const textBox = this._postsContainer.querySelector("form");

    const message = document.createElement("textarea");
    message.id = "new-message";
    message.name = "message";
    message.placeholder = "Type your message here...";
    textBox.append(message);
    let shiftPressed = false;

    message?.addEventListener("keydown", (event) => {
      // // console.log("key is " + event.key);
      if (event.key == "Shift") {
        shiftPressed = true;
      }

      if (event.key == "Enter") {
        // // console.log("Enter pressed");
        if (!shiftPressed) {
          // // console.log("Showing post input");
          const messageEvent = new CustomEvent("messageEvent", {
            detail: {
              name: this._openChannelId,
              message: message.value,
              parent: this._parent,
            },
          });
          document.dispatchEvent(messageEvent);
        } else {
          // // console.log("Enter shift");
        }
      }
    });

    message?.addEventListener("keyup", (event) => {
      if (event.key == "Shift") {
        shiftPressed = false;
      }
    });

    const sendButton = document.createElement("button");
    sendButton.type = "button";
    sendButton.id = "send-button";
    sendButton.textContent = "Send";
    textBox.append(sendButton);

    sendButton?.addEventListener("click", () => {
      // console.log("Send clicked");
      // console.log("Showing post input");

      const messageEvent = new CustomEvent("messageEvent", {
        detail: {
          name: this._openChannelId,
          message: message.value,
          parent: this._parent,
        },
      });

      document.dispatchEvent(messageEvent);
    });

    const smileButton = document.createElement("button");
    smileButton.type = "button";
    smileButton.id = "smile-button";
    const smileIcon = document.createElement("iconify-icon");
    smileIcon.setAttribute("icon", "emojione:smiling-face");
    smileIcon.setAttribute("width", "1.25em");
    smileIcon.setAttribute("height", "1.25em");
    smileButton.append(smileIcon);
    textBox.append(smileButton);

    const smileDesc = document.createElement("p");
    smileDesc.id = "smile-desc";
    smileDesc.textContent = "Add a smile reaction to your post";
    smileDesc.style.display = "none"; // Hide the description from view but still accessible to screen readers
    smileButton.setAttribute("aria-describedby", "smile-desc");
    smileButton.setAttribute("aria-label", "Add a smile reaction to post");

    smileButton?.addEventListener("click", () => {
      // console.log("Smile clicked");
      message.value = message.value.concat(":smile:");
    });

    const frownButton = document.createElement("button");
    frownButton.type = "button";
    frownButton.id = "frown-button";
    const frownIcon = document.createElement("iconify-icon");
    frownIcon.setAttribute("icon", "emojione:frowning-face");
    frownIcon.setAttribute("width", "1.25em");
    frownIcon.setAttribute("height", "1.25em");
    frownButton.append(frownIcon);
    textBox.append(frownButton);

    // Descriptive text for the smile button
    const frownDesc = document.createElement("p");
    frownDesc.id = "frown-desc";
    frownDesc.textContent = "Add a frown reaction to your post";
    frownDesc.style.display = "none"; // Hide the description from view but still accessible to screen readers
    frownButton.setAttribute("aria-describedby", "frown-desc");
    frownButton.setAttribute("aria-label", "React with a frown");

    frownButton?.addEventListener("click", () => {
      // console.log("Frown clicked");
      message.value = message.value.concat(":frown:");
    });

    const likeButton = document.createElement("button");
    likeButton.type = "button";
    likeButton.id = "like-button";
    const likeIcon = document.createElement("iconify-icon");
    likeIcon.setAttribute("icon", "emojione:thumbs-up");
    likeIcon.setAttribute("width", "1.25em");
    likeIcon.setAttribute("height", "1.25em");
    likeButton.append(likeIcon);
    textBox.append(likeButton);

    likeButton?.addEventListener("click", () => {
      // console.log("Like clicked");
      message.value = message.value.concat(":like:");
    });

    const likeDesc = document.createElement("p");
    likeDesc.id = "like-desc";
    likeDesc.textContent = "Add a like in your post";
    likeDesc.style.display = "none"; // Hide the description from view but still accessible to screen readers
    likeButton.setAttribute("aria-describedby", "like-desc");
    likeButton.setAttribute("aria-label", "React with a like");

    const celebrateButton = document.createElement("button");
    celebrateButton.type = "button";
    celebrateButton.id = "celebrate-button";
    const celebrateIcon = document.createElement("iconify-icon");
    celebrateIcon.setAttribute("icon", "emojione:party-popper");
    celebrateIcon.setAttribute("width", "1.25em");
    celebrateIcon.setAttribute("height", "1.25em");
    celebrateButton.append(celebrateIcon);
    textBox.append(celebrateButton);

    celebrateButton?.addEventListener("click", () => {
      // console.log("Celebrate clicked");
      message.value = message.value.concat(":celebrate:");
    });

    const CelebrateDesc = document.createElement("p");
    CelebrateDesc.id = "celebrate-desc";
    CelebrateDesc.textContent = "Add a celebration emoji to your post";
    CelebrateDesc.style.display = "none"; // Hide the description from view but still accessible to screen readers
    celebrateButton.setAttribute("aria-describedby", "celebrate-desc");

    celebrateButton.setAttribute(
      "aria-label",
      "Add a celebration emoji to post",
    );

    const boldButton = document.createElement("button");
    boldButton.type = "button";
    boldButton.id = "bold-button";
    boldButton.textContent = "Bold";
    boldButton.style.fontWeight = "bold";
    textBox.append(boldButton);

    const boldDesc = document.createElement("p");
    boldDesc.id = "bold-desc";
    boldDesc.textContent = "set bold style to highlighted text";
    boldDesc.style.display = "none"; // Hide the description from view but still accessible to screen readers
    boldButton.setAttribute("aria-describedby", "bold-desc");
    boldButton.setAttribute("aria-label", "set bold style to highlighted text");

    boldButton?.addEventListener("click", () => {
      // console.log("Bold clicked");
      // console.log("Selected text index start:", message.selectionStart);
      // console.log("Selected text index end:", message.selectionEnd);
      message.value = message.value
        .substring(0, message.selectionStart)
        .concat(
          "**",
          message.value.substring(message.selectionStart, message.selectionEnd),
          "**",
          message.value.substring(message.selectionEnd),
        );
    });

    const italicButton = document.createElement("button");
    italicButton.type = "button";
    italicButton.id = "italic-button";
    italicButton.textContent = "Italic";
    italicButton.style.fontStyle = "italic";
    textBox.append(italicButton);

    const italicDesc = document.createElement("p");
    italicDesc.id = "italic-desc";
    italicDesc.textContent = "set italic style to highlighted text";
    italicDesc.style.display = "none"; // Hide the description from view but still accessible to screen readers
    italicButton.setAttribute("aria-describedby", "italic-desc");

    italicButton?.addEventListener("click", () => {
      // console.log("Italic clicked");
      message.value = message.value
        .substring(0, message.selectionStart)
        .concat(
          "*",
          message.value.substring(message.selectionStart, message.selectionEnd),
          "*",
          message.value.substring(message.selectionEnd),
        );
    });

    const linkButton = document.createElement("button");
    linkButton.type = "button";
    linkButton.id = "link-button";
    linkButton.textContent = "Link";
    linkButton.style.textDecoration = "underline";
    textBox.append(linkButton);

    linkButton?.addEventListener("click", () => {
      // console.log("Link clicked");
      message.value = message.value
        .substring(0, message.selectionStart)
        .concat(
          "[",
          message.value.substring(message.selectionStart, message.selectionEnd),
          "]",
          "(",
          message.value.substring(message.selectionStart, message.selectionEnd),
          ")",
          message.value.substring(message.selectionEnd),
        );
    });

    const superscriptButton = document.createElement("button");
    superscriptButton.type = "button";
    superscriptButton.id = "superscript-button";
    superscriptButton.innerHTML = "Super<sup>script</sup>";
    textBox.append(superscriptButton);

    const superDesc = document.createElement("p");
    superDesc.id = "super-desc";
    superDesc.textContent = "set italic style to highlighted text";
    superDesc.style.display = "none"; // Hide the description from view but still accessible to screen readers
    superscriptButton.setAttribute("aria-describedby", "super-desc");

    superscriptButton?.addEventListener("click", () => {
      // console.log("Superscript clicked");
      message.value = message.value
        .substring(0, message.selectionStart)
        .concat(
          "^^",
          message.value.substring(message.selectionStart, message.selectionEnd),
          "^^",
          message.value.substring(message.selectionEnd),
        );
    });

    const subscriptButton = document.createElement("button");
    subscriptButton.type = "button";
    subscriptButton.id = "subscript-button";
    subscriptButton.innerHTML = "Sub<sub>script</sub>";
    textBox.append(subscriptButton);
    const subDesc = document.createElement("p");
    subDesc.id = "sub-desc";
    subDesc.textContent = "set italic style to highlighted text";
    subDesc.style.display = "none"; // Hide the description from view but still accessible to screen readers
    subscriptButton.setAttribute("aria-describedby", "sub-desc");

    subscriptButton?.addEventListener("click", () => {
      // console.log("Subscript clicked");
      message.value = message.value
        .substring(0, message.selectionStart)
        .concat(
          "__",
          message.value.substring(message.selectionStart, message.selectionEnd),
          "__",
          message.value.substring(message.selectionEnd),
        );
    });
  }

  /**
   * Display all posts in Channel
   *
   * @param posts array of workspaces (documents in the database) to display
   */
  displayPosts(posts: Array<ViewPost>, channel: string): void {
    // console.log("Posts displaying");
    this._sending = false;
    var hierarhcyPosts = this.makePostHierarhcy(posts);
    if (this._openChannelId === channel) {
      // console.log("Channel is already open.");
      return;
    } else {
      // If a different channelis open, close all posts from the previous one
      if (this._openChannelId) {
        const prevChannel = this._channelsContainer.querySelector(
          `[data-channel-id="${this._openChannelId}"]`,
        );
        if (prevChannel) {
          prevChannel.DeactivateState();
        }

        this.closeAllPosts();
      }
    }

    this.showPosts(hierarhcyPosts, channel);
  }

  /**
   * Display all posts in Channel
   *
   * @param posts array of workspaces (documents in the database) to display
   */
  updatePosts(posts: Array<ViewPost>, channel: string): void {
    // console.log("Posts UPDATING");

    // if (this._hierarhcyPosts) {
    //   // console.log("just to use hierarchyPosts");
    // }
    this._sending = false;
    var hierarhcyPosts = this.makePostHierarhcy(posts);

    this.closeAllPosts();

    this.showPosts(hierarhcyPosts, channel);
  }

  /**
   * Scroll to current post
   *
   * @param currentPost post to be scrolled into view
   */
  scroll(currentPost: string): void {
    console.log("scroll into view");
    const postElement = this._postsContainer?.querySelector(
      `[data-post-id="${currentPost}"]`,
    );
    if (postElement) {
      postElement.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  }

  /**
   * close all posts in currenlty open channel
   *
   */
  closeAllPosts(): void {
    // Removing current header if there is one
    const oldheader = this._postsContainer?.querySelector("h2");
    oldheader?.remove();
    this._openPosts.forEach((postId) => {
      // go through the channels and remove each
      const postElement = this._postsContainer?.querySelector(
        `[data-post-id="${postId}"]`,
      );
      if (postElement) {
        postElement.remove();
      }
    });

    const currerror = this._postsContainer?.querySelector(`[id = "currError"]`);

    if (currerror) {
      currerror.remove();
    }

    const message = this._postsContainer?.querySelector("textarea");
    if (message) {
      message.remove();
    }
    let buttonsLeft = true;
    while (buttonsLeft) {
      const button = this._postsContainer?.querySelector("button");
      if (button) {
        button.remove();
      } else {
        buttonsLeft = false;
      }
    }
    this._openChannelId = null;
    this._parent = undefined;
  }

  /**
   * Display an error message to the user
   *
   * @param error error message to display
   */
  displayError(error: string): void {
    // console.log("displaying error!!!");
    const modal = document.createElement("section");
    modal.style.position = "fixed";
    modal.style.left = "0";
    modal.style.top = "0";
    modal.style.width = "100%";
    modal.style.height = "100%";
    modal.style.backgroundColor = "rgba(0, 0, 0, 0.5)";
    modal.style.display = "flex";
    modal.style.justifyContent = "center";
    modal.style.alignItems = "center";
    modal.style.zIndex = "1000";

    // Create the modal content
    const modalContent = document.createElement("section");
    modalContent.style.backgroundColor = "#fff";
    modalContent.style.padding = "20px";
    modalContent.style.borderRadius = "5px";
    modalContent.style.textAlign = "center";

    const errorMessage = document.createElement("p");
    errorMessage.textContent = error;
    errorMessage.setAttribute("aria-live", "assertive");
    modalContent.appendChild(errorMessage);

    const dismissButton = document.createElement("button");
    dismissButton.textContent = "Dismiss";
    dismissButton.style.marginTop = "10px";

    const Desc = document.createElement("p");
    Desc.id = "desc";
    Desc.textContent = "dismiss error button";
    Desc.style.display = "none";
    dismissButton.setAttribute("aria-describedby", "sub-desc");
    dismissButton.onclick = function () {
      modal.remove();
    };

    modalContent.appendChild(dismissButton);

    modal.appendChild(modalContent);

    document.body.appendChild(modal);
  }
}
