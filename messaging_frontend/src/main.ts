import { slog } from "./slog";
import { Model } from "./model";
import { View } from "./view";

import { Doc } from "./types";

import { Validator } from "./validation";

import {
  LoginEvent,
  LogoutEvent,
  openEvent,
  openEventChannel,
  refreshChannels,
  refreshWorkspaces,
  ViewPost,
  DeleteChannel,
  ViewWorkspace,
  DeleteWorkspace,
  messageEvent,
  replytoPost,
  jsonDoc,
  reaction,
  patchEvent,
  createChannelEvent,
  createWorkspaceEvent,
  subscribeEvent,
  abortEvent,
} from "./types";

/**
 * Inital entry to point of the application.
 */
function main(): void {
  slog.info("Using database", [
    "database",
    `${process.env.DATABASE_HOST}${process.env.DATABASE_PATH}`,
  ]);
  new Controller(); // Instantiate the App class
}

document.addEventListener("DOMContentLoaded", main);

/**
 * Class Controller
 */
class Controller {
  model: Model;
  view: View;
  isLoggedIn: boolean;
  validator: Validator;
  abortControl: AbortController = new AbortController();

  /**
   * Declare names and types of environment variables.
   */
  declare process: {
    env: {
      DATABASE_HOST: string;
      DATABASE_PATH: string;
      AUTH_PATH: string;
    };
  };

  constructor() {
    slog.info("Using database", [
      "database",
      `${process.env.DATABASE_HOST}${process.env.DATABASE_PATH}`,
    ]);

    this.model = new Model();
    this.view = new View();
    this.isLoggedIn = false;
    this.validator = new Validator();
    this.registerEventListeners();
    this.updateView(null);
  }

  /**
   * Updates the View depending on if a user is logged in or not.
   *
   * @param username the username of the person who is currenlty logged in
   */
  updateView(username: string | null): void {
    if (this.isLoggedIn) {
      this.view.displayPage(username || "");
      this.model
        .getWorkspaces(this.validator) // gets all the documents at path definied in .env
        .then((response: Array<Doc>) => {
          console.log("response from geting workspaces", response);
          this.view.displayAllWorkspaces(
            response as unknown as Array<ViewWorkspace>,
            false,
          );
        })
        .catch((error) => {
          console.log(error);
        });
    } else {
      console.log("calling display login promt");
      // User is not logged in, display login
      this.view.displayLoginPrompt(); // Implement this method to display a login form
    }
  }

  /**
   * Assigns all event listeners nedded for the set up of the Website
   */
  registerEventListeners() {
    document.addEventListener("loginEvent", this.handleLoginEvent.bind(this));
    document.addEventListener("logoutEvent", this.handleLogoutEvent.bind(this));
    document.addEventListener(
      "refreshWorkspaces",
      this.handleRefreshWorkspacesEvent.bind(this),
    );
    document.addEventListener("openEvent", this.handleopenEvent.bind(this));
    document.addEventListener(
      "deleteWorkspace",
      this.handledeleteWorkspace.bind(this),
    );
    document.addEventListener(
      "deleteChannel",
      this.handledeleteChannel.bind(this),
    );
    document.addEventListener(
      "openEventChannel",
      this.handleopenChannelEvent.bind(this),
    );
    document.addEventListener(
      "createChannelEvent",
      this.handlecreateChannelEvent.bind(this),
    );
    document.addEventListener(
      "createWorkspaceEvent",
      this.handlecreateWorkspaceEvent.bind(this),
    );
    document.addEventListener(
      "refreshChannels",
      this.refreshChannelsEvent.bind(this),
    );
    document.addEventListener("abortEvent", this.abortEvent.bind(this));
    document.addEventListener("replytoPost", this.replytoPostEvent.bind(this));
    document.addEventListener(
      "messageEvent",
      this.handlemessageEvent.bind(this),
    );
    document.addEventListener(
      "subscribeEvent",
      this.handlesubscribeEvent.bind(this),
    );

    document.addEventListener("patchEvent", this.handlepatchEvent.bind(this));
  }

  /**
   * Handles the processing of a newPostEvent event.
   *
   * This function is activated upon receiving a `newPostEvent`.
   * It creating a new reaction object to track reactions to the post and
   * forms the JSON document to be stored in the database, the post data is delegated thorugh
   * the newPostEvent evnt detail
   * after the document has been stored in the database it calls the posts to be updated in the view
   *
   *
   * @param {CustomEvent<newPostEvent>} evt - The event carrying the message and related data.
   */

  handlemessageEvent(evt: CustomEvent<messageEvent>) {
    console.log("messageEvent", evt);
    console.log("message we get", evt.detail.message);

    const newReaction: reaction = {
      smile: [],
      frown: [],
      like: [],
      celebrate: [],
    };
    const newJsonDoc: jsonDoc = {
      msg: evt.detail.message,
      parent: evt.detail.parent,
      reactions: newReaction,
      extensions: {},
    };

    // First, put the post in the database
    this.model
      .putPostInDB(evt.detail.name, newJsonDoc)
      .then((response) => {
        console.log("response from putting a post in db", response);
        // Only after the post is successfully put in the database, open the channel
        return this.model.openChannel(evt.detail.name, this.validator);
      })
      .then((response: Array<Doc>) => {
        console.log("posts from channel", response);
        var posts = response as unknown as Array<ViewPost>;
        this.view.updatePosts(posts, evt.detail.name);
      })
      .then(() => {
        this.view.scroll(evt.detail.parent);
      })
      .catch((error) => {
        // Handle any errors that occur during the process
        console.log(error);
        this.view.displayError(error);
      });
  }

  /**
   * Handles the processing of a patchEvent event and
   * delegeates information to the model to send a patch request with the given details
   *
   *
   * @param {CustomEvent<patchEvent>} evt - The event carrying data for for the patch
   */

  handlepatchEvent(evt: CustomEvent<patchEvent>) {
    console.log(evt);

    const name = this.view.getUsername();
    this.model.updateReaction(
      evt.detail.op,
      evt.detail.reaction,
      name,
      evt.detail.path,
    );
  }

  /*
   * Handles the replytoPost event, sets parent of post
   *
   *  @param {CustomEvent<replytoPost>} evt
   */
  replytoPostEvent(evt: CustomEvent<replytoPost>) {
    console.log("replytoPost", evt);
    console.log("now replying to", evt.detail.id);
    this.view.setParent(evt.detail.id);
  }
  handlesubscribeEvent(evt: CustomEvent<subscribeEvent>) {
    let updatedPost = evt.detail.data as ViewPost;

    this.model
      .openChannel(
        updatedPost.path.substring(0, updatedPost.path.indexOf("/posts")),
        this.validator,
      )
      .then((response: Array<Doc>) => {
        console.log("posts from channel", response);
        var posts = response as unknown as Array<ViewPost>;
        this.view.updatePosts(
          posts,
          updatedPost.path.substring(0, updatedPost.path.indexOf("/posts")),
        );
      })
      .catch((error) => {
        // Handle any errors that occur during the process
        console.log(error);
        this.view.displayError(error);
      });
  }

  abortEvent(evt: CustomEvent<abortEvent>) {
    if (!evt.detail.abort) {
      console.log("problem here...");
    }
    console.log("abortEvent occurs");
    console.log("Cancelling subscription");
    this.abortControl.abort();
    this.abortControl = new AbortController();
  }

  /**
   * Handles the refreshChannels event, delegates the model and view to refresh channels
   *
   * @param {CustomEvent<refreshChannels>} evt
   */
  refreshChannelsEvent(evt: CustomEvent<refreshChannels>) {
    console.log("refreshChannels", evt);
    this.model
      .openWorkspace(evt.detail.name, this.validator)
      .then((response: Array<Doc>) => {
        console.log("response from opening a single workspace", response);
        this.view.displayChannels(
          response as unknown as Array<ViewWorkspace>,
          evt.detail.name,
          true,
        );
      })
      .catch((error) => {
        console.log(error);
        this.view.dispalyChannelError(error);
      });
  }

  /**
   * Handle the createWorkspaceEvent and calls model to create a workspace
   * in owlDB then calling an update on view to represent the new changes
   *
   * @param {CustomEvent<createWorkspaceEvent>} evt
   */
  handlecreateWorkspaceEvent(evt: CustomEvent<createWorkspaceEvent>) {
    console.log("createWorkspaceEvent", evt);
    this.model
      .createWorkspace(evt.detail.workspace)
      .then((response) => {
        console.log("response from creating a channel", response);
        return this.model.getWorkspaces(this.validator);
      })
      .then((response) => {
        console.log("response from getting workspaces", response);
        this.view.displayAllWorkspaces(
          response as unknown as Array<ViewWorkspace>,
          true,
        );
        return this.model.openWorkspace(
          "/" + evt.detail.workspace,
          this.validator,
        );
      })
      .then((response) => {
        console.log("response from opening a single workspace", response);
        this.view.displayChannels(
          response as unknown as Array<ViewWorkspace>,
          "/" + evt.detail.workspace,
          false,
        );
      })
      .catch((error) => {
        console.log("Got an error in creating a workspace");
        console.log(error);
        this.view.dispalyWorkspaceError(error);
      });
  }

  /**
   * Handles the opening of a channel based on the OpenEventChannel event.
   *
   * The function calls the function in model to retrieve the channels from the
   * database and then calls the view to dispaly the channels
   *
   * @param {CustomEvent<openChannelEvent>} evt custom event indicting channel that should be opened
   */
  handleopenChannelEvent(evt: CustomEvent<openEventChannel>) {
    console.log("openChannelEvent", evt);

    console.log("abortEvent occurs");
    console.log("Cancelling subscription");
    this.abortControl.abort();
    this.abortControl = new AbortController();

    this.model
      .openChannel(evt.detail.name, this.validator)
      .then((response: Array<Doc>) => {
        console.log("response from opening a single channel", response);
        this.view.displayPosts(
          response as unknown as Array<ViewPost>,
          evt.detail.name,
        );
        this.model.startSubscription(
          `${process.env.DATABASE_HOST}${process.env.DATABASE_PATH}` +
            evt.detail.name,
          this.abortControl,
        );
      })
      .catch((error) => {
        console.log(error);
        this.view.dispalyChannelError(error);
      });
  }

  /**
   * Handles the reateChannelEvent and calls model to create a channel
   * in owlDB then calling an update on view to represent the new changes
   *
   * @param {CustomEvent<createChannelEvent>} evt
   */
  handlecreateChannelEvent(evt: CustomEvent<createChannelEvent>) {
    console.log("createChannelEvent", evt);
    this.model
      .createChannel(evt.detail.workspace, evt.detail.channel)
      .then((response) => {
        console.log("response from creating a channel", response);
        return this.model.openWorkspace(evt.detail.workspace, this.validator);
      })
      .then((response: Array<Doc>) => {
        console.log("response from opening a single workspace", response);
        this.view.displayChannels(
          response as unknown as Array<ViewWorkspace>,
          evt.detail.workspace,
          true,
        );
      })
      .catch((error) => {
        console.log(error);
        this.view.dispalyChannelError(error);
        // this.view.displayError(error);
      });
  }

  /**
   * Handles the deletion of a channel based on the DeleteChannel event.
   *
   * This function is triggered by an event that contains details about the channel to be deleted.
   * It gets the path of the channel from the DeleteChannel event detail and calls the model's deleteChannel method
   * to perform the deletion. After the channel is succesfully deleted, it opens the workspace containing the
   * channel and updates the view to reflect the changes.
   *
   * The function handles the entire delete operation including catching and displaying any errors
   * that might occur during the process.
   *
   * @param {CustomEvent<DeleteChannel>} evt custom event indicting  the path to which channel should be deleted
   */
  handledeleteChannel(evt: CustomEvent<DeleteChannel>) {
    console.log("deleteChannel", evt);
    const parts = evt.detail.id.split("/");
    this.model
      .deleteChannel(evt.detail.id)

      .then((response) => {
        // Create Channels
        console.log("response from deleting a a channel", response);

        return this.model.openWorkspace("/" + parts[1], this.validator);
      })
      .then((response: Array<Doc>) => {
        console.log("response from opening a single workspace", response);
        this.view.displayChannels(
          response as unknown as Array<ViewWorkspace>,
          "/" + parts[1],
          true,
        );
      })
      .catch((error) => {
        console.log(error);
        this.view.dispalyChannelError(error);
        // this.view.displayError(error);
      });
  }

  /**
   * Handles the deletion of a workspace based on the DeleteWorkpace event.
   *
   * This function is triggered by an event that contains details about the workspace to be deleted.
   * It calls the model's deleteWorkspace method to perform the deletion. After the workspace is
   * succesfully deleted, it opens the refreshes the workspaces to reflect the changes
   *
   * The function handles the entire delete operation including catching and displaying any errors
   * that might occur during the process.
   *
   * @param {CustomEvent<DeleteWorkspace>} evt custom event indicting the path to which workpace should be deleted
   */
  handledeleteWorkspace(evt: CustomEvent<DeleteWorkspace>) {
    console.log("deleteWorkspace", evt);
    this.model
      .deleteWorkspace(evt.detail.id)
      .then(() => {
        const workspaceDeletedEvent = new CustomEvent("workspaceDeleted", {
          detail: { id: evt.detail.id },
        });
        document.dispatchEvent(workspaceDeletedEvent);

        const refreshWorkspaces = new CustomEvent("refreshWorkspaces", {
          detail: { name: "path" },
        });
        console.log("refreshing workspaces");
        document.dispatchEvent(refreshWorkspaces);
      })
      .catch((error) => {
        console.log(error);
        // this.view.displayError(error);
        this.view.dispalyWorkspaceError(error); //TODO by trying to delete a workspace thats already deleted
      });
  }

  /**
   * Handles the openEvent event which delegates the opening
   * of workspace, calls the model and view to adjust acordingly
   *
   * @param {CustomEvent<openEvent>} evt
   */
  handleopenEvent(evt: CustomEvent<openEvent>) {
    console.log("abortEvent occurs");
    console.log("Cancelling subscription");
    this.abortControl.abort();
    this.abortControl = new AbortController();
    console.log("openEvent", evt);
    this.model
      .openWorkspace(evt.detail.name, this.validator)
      .then((response: Array<Doc>) => {
        console.log("response from opening a single workspace", response);
        this.view.displayChannels(
          response as unknown as Array<ViewWorkspace>,
          evt.detail.name,
          false,
        );
      })
      .catch((error) => {
        console.log(error);
        this.view.dispalyWorkspaceError(error);
      });
  }

  /**
   * Handles the LogoutEvent, delegates the function of logging out a
   * user to the model and view, updating the view to be ready to login again
   *
   * @param {CustomEvent<LogoutEvent>} evt
   */
  handleLogoutEvent(evt: CustomEvent<LogoutEvent>) {
    console.log("abortEvent occurs");
    console.log("Cancelling subscription");
    this.abortControl.abort();
    this.abortControl = new AbortController();

    this.model
      .logout() // call authenticate with username
      .then(() => {
        this.isLoggedIn = false;
        this.updateView(null);
      })
      .catch((error) => {
        console.log(error);
        this.view.displayError(`Logout error:  ${error}`);
      });
  }

  /**
   * Handles the event refreshWorkspaces, delegates the function of refreshing a workspace to
   * the model and view and updating the view to represent the refreshed workspaces
   *
   * @param {CustomEvent<refreshWorkspaces>} evt custom event for refreshing a workspace
   */
  handleRefreshWorkspacesEvent(evt: CustomEvent<refreshWorkspaces>) {
    this.abortControl.abort();
    this.abortControl = new AbortController();
    this.model
      .getWorkspaces(this.validator) // gets all the documents at path definied in .env
      .then((response: Array<Doc>) => {
        console.log("response from geting workspaces", response);
        this.view.displayAllWorkspaces(
          response as unknown as Array<ViewWorkspace>,
          true,
        );
      })
      .catch((error) => {
        console.log(error);
        this.view.displayError(error);
      });
  }

  /**
   *Handles the login event dispatched upon a login atempt. Delegates the username to the model to authenticate it.
   *
   * @param {CustomEvent<LoginEvent>} evt
   */
  handleLoginEvent(evt: CustomEvent<LoginEvent>) {
    const username = evt.detail.message;
    console.log("username", username);

    if (this.isLoggedIn) {
      console.log("already logged in");
      return;
    } else {
      console.log("not logged in");
      this.model
        .authenticate(username)
        .then(() => {
          this.isLoggedIn = true;
          this.updateView(username);
        })
        .catch((error) => {
          console.log(error);
          this.view.displayLoginError(error);
        });
    }
  }
}

export { Controller };
